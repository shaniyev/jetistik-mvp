
-- =============================================
-- JETISTIK.KZ MVP — Phase 1: Core Schema
-- =============================================

-- 1. Role enum
CREATE TYPE public.app_role AS ENUM ('student', 'parent', 'coach', 'admin');

-- 2. Profiles table
CREATE TABLE public.profiles (
  id uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  role app_role NOT NULL DEFAULT 'student',
  full_name text NOT NULL DEFAULT '',
  grade text,
  school text,
  city text,
  is_public boolean NOT NULL DEFAULT false,
  public_token text UNIQUE DEFAULT encode(gen_random_bytes(16), 'hex'),
  parent_invite_code text UNIQUE DEFAULT upper(substr(encode(gen_random_bytes(4), 'hex'), 1, 8)),
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

-- 3. User roles table (separate from profile role for security)
CREATE TABLE public.user_roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  role app_role NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(user_id, role)
);
ALTER TABLE public.user_roles ENABLE ROW LEVEL SECURITY;

-- 4. Parent links
CREATE TABLE public.parent_links (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  student_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(parent_id, student_id)
);
ALTER TABLE public.parent_links ENABLE ROW LEVEL SECURITY;

-- 5. Subjects
CREATE TABLE public.subjects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  sort_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.subjects ENABLE ROW LEVEL SECURITY;

-- 6. Topics
CREATE TABLE public.topics (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id uuid NOT NULL REFERENCES public.subjects(id) ON DELETE CASCADE,
  name text NOT NULL,
  sort_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.topics ENABLE ROW LEVEL SECURITY;

-- 7. Questions
CREATE TABLE public.questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  topic_id uuid NOT NULL REFERENCES public.topics(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('mcq', 'short')),
  prompt text NOT NULL,
  choices jsonb,
  correct_answer text NOT NULL,
  explanation text,
  difficulty int NOT NULL DEFAULT 1 CHECK (difficulty BETWEEN 1 AND 5),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.questions ENABLE ROW LEVEL SECURITY;

-- 8. Attempts
CREATE TABLE public.attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  subject_id uuid NOT NULL REFERENCES public.subjects(id),
  topic_id uuid NOT NULL REFERENCES public.topics(id),
  mode text NOT NULL CHECK (mode IN ('training', 'test')) DEFAULT 'training',
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  score int DEFAULT 0,
  total int DEFAULT 0,
  duration_seconds int
);
ALTER TABLE public.attempts ENABLE ROW LEVEL SECURITY;

-- 9. Attempt answers
CREATE TABLE public.attempt_answers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id uuid NOT NULL REFERENCES public.attempts(id) ON DELETE CASCADE,
  question_id uuid NOT NULL REFERENCES public.questions(id),
  answer text,
  is_correct boolean NOT NULL DEFAULT false
);
ALTER TABLE public.attempt_answers ENABLE ROW LEVEL SECURITY;

-- 10. Groups
CREATE TABLE public.groups (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  coach_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  subject_id uuid REFERENCES public.subjects(id),
  title text NOT NULL,
  invite_code text UNIQUE DEFAULT upper(substr(encode(gen_random_bytes(4), 'hex'), 1, 8)),
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.groups ENABLE ROW LEVEL SECURITY;

-- 11. Group members
CREATE TABLE public.group_members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES public.groups(id) ON DELETE CASCADE,
  student_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  joined_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(group_id, student_id)
);
ALTER TABLE public.group_members ENABLE ROW LEVEL SECURITY;

-- 12. Certificates
CREATE TABLE public.certificates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text UNIQUE NOT NULL DEFAULT upper(substr(encode(gen_random_bytes(6), 'hex'), 1, 12)),
  student_name text NOT NULL,
  event text NOT NULL,
  result text,
  date date NOT NULL DEFAULT CURRENT_DATE,
  issuer text,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.certificates ENABLE ROW LEVEL SECURITY;

-- =============================================
-- SECURITY DEFINER FUNCTIONS
-- =============================================

CREATE OR REPLACE FUNCTION public.has_role(_user_id uuid, _role app_role)
RETURNS boolean
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = public
AS $$
  SELECT EXISTS (
    SELECT 1 FROM public.user_roles
    WHERE user_id = _user_id AND role = _role
  )
$$;

CREATE OR REPLACE FUNCTION public.is_admin(_user_id uuid)
RETURNS boolean
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = public
AS $$
  SELECT public.has_role(_user_id, 'admin')
$$;

-- Trigger to create profile + role on signup
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  _role app_role;
  _full_name text;
BEGIN
  _role := COALESCE((NEW.raw_user_meta_data->>'role')::app_role, 'student');
  _full_name := COALESCE(NEW.raw_user_meta_data->>'full_name', '');

  INSERT INTO public.profiles (id, role, full_name)
  VALUES (NEW.id, _role, _full_name);

  INSERT INTO public.user_roles (user_id, role)
  VALUES (NEW.id, _role);

  RETURN NEW;
END;
$$;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- =============================================
-- RLS POLICIES
-- =============================================

-- Profiles
CREATE POLICY "Users can view own profile" ON public.profiles FOR SELECT USING (auth.uid() = id);
CREATE POLICY "Users can update own profile" ON public.profiles FOR UPDATE USING (auth.uid() = id);
CREATE POLICY "Admin full access profiles" ON public.profiles FOR ALL USING (public.is_admin(auth.uid()));
CREATE POLICY "Public profiles viewable" ON public.profiles FOR SELECT USING (is_public = true);

-- User roles
CREATE POLICY "Users can view own roles" ON public.user_roles FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Admin full access roles" ON public.user_roles FOR ALL USING (public.is_admin(auth.uid()));

-- Parent links
CREATE POLICY "Parents and students can view own links" ON public.parent_links FOR SELECT USING (auth.uid() = parent_id OR auth.uid() = student_id);
CREATE POLICY "Parents can create links" ON public.parent_links FOR INSERT WITH CHECK (auth.uid() = parent_id);
CREATE POLICY "Admin full access parent_links" ON public.parent_links FOR ALL USING (public.is_admin(auth.uid()));

-- Subjects (public read, admin write)
CREATE POLICY "Anyone can read subjects" ON public.subjects FOR SELECT USING (true);
CREATE POLICY "Admin can manage subjects" ON public.subjects FOR ALL USING (public.is_admin(auth.uid()));

-- Topics (public read, admin write)
CREATE POLICY "Anyone can read topics" ON public.topics FOR SELECT USING (true);
CREATE POLICY "Admin can manage topics" ON public.topics FOR ALL USING (public.is_admin(auth.uid()));

-- Questions (authenticated read, admin write)
CREATE POLICY "Authenticated can read active questions" ON public.questions FOR SELECT TO authenticated USING (is_active = true);
CREATE POLICY "Admin can manage questions" ON public.questions FOR ALL USING (public.is_admin(auth.uid()));

-- Attempts
CREATE POLICY "Students can create attempts" ON public.attempts FOR INSERT WITH CHECK (auth.uid() = student_id);
CREATE POLICY "Students can view own attempts" ON public.attempts FOR SELECT USING (auth.uid() = student_id);
CREATE POLICY "Students can update own attempts" ON public.attempts FOR UPDATE USING (auth.uid() = student_id);
CREATE POLICY "Coach can view group student attempts" ON public.attempts FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.group_members gm
    JOIN public.groups g ON g.id = gm.group_id
    WHERE gm.student_id = attempts.student_id AND g.coach_id = auth.uid()
  )
);
CREATE POLICY "Parent can view linked student attempts" ON public.attempts FOR SELECT USING (
  EXISTS (SELECT 1 FROM public.parent_links WHERE parent_id = auth.uid() AND student_id = attempts.student_id)
);
CREATE POLICY "Admin full access attempts" ON public.attempts FOR ALL USING (public.is_admin(auth.uid()));

-- Attempt answers
CREATE POLICY "Students can insert own answers" ON public.attempt_answers FOR INSERT WITH CHECK (
  EXISTS (SELECT 1 FROM public.attempts WHERE id = attempt_answers.attempt_id AND student_id = auth.uid())
);
CREATE POLICY "Students can view own answers" ON public.attempt_answers FOR SELECT USING (
  EXISTS (SELECT 1 FROM public.attempts WHERE id = attempt_answers.attempt_id AND student_id = auth.uid())
);
CREATE POLICY "Coach can view group answers" ON public.attempt_answers FOR SELECT USING (
  EXISTS (
    SELECT 1 FROM public.attempts a
    JOIN public.group_members gm ON gm.student_id = a.student_id
    JOIN public.groups g ON g.id = gm.group_id
    WHERE a.id = attempt_answers.attempt_id AND g.coach_id = auth.uid()
  )
);
CREATE POLICY "Admin full access answers" ON public.attempt_answers FOR ALL USING (public.is_admin(auth.uid()));

-- Groups
CREATE POLICY "Coach can manage own groups" ON public.groups FOR ALL USING (auth.uid() = coach_id);
CREATE POLICY "Members can view their groups" ON public.groups FOR SELECT USING (
  EXISTS (SELECT 1 FROM public.group_members WHERE group_id = groups.id AND student_id = auth.uid())
);
CREATE POLICY "Admin full access groups" ON public.groups FOR ALL USING (public.is_admin(auth.uid()));

-- Group members
CREATE POLICY "Coach can manage group members" ON public.group_members FOR ALL USING (
  EXISTS (SELECT 1 FROM public.groups WHERE id = group_members.group_id AND coach_id = auth.uid())
);
CREATE POLICY "Students can view own memberships" ON public.group_members FOR SELECT USING (auth.uid() = student_id);
CREATE POLICY "Students can join groups" ON public.group_members FOR INSERT WITH CHECK (auth.uid() = student_id);
CREATE POLICY "Admin full access group_members" ON public.group_members FOR ALL USING (public.is_admin(auth.uid()));

-- Certificates (public read for active, admin write)
CREATE POLICY "Public can view active certificates" ON public.certificates FOR SELECT USING (active = true);
CREATE POLICY "Admin can manage certificates" ON public.certificates FOR ALL USING (public.is_admin(auth.uid()));
