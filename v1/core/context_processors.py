def user_role_flags(request):
    user = getattr(request, "user", None)
    if not user or not getattr(user, "is_authenticated", False):
        return {"is_student": False, "is_teacher": False}
    return {
        "is_student": user.groups.filter(name="user_student").exists(),
        "is_teacher": user.groups.filter(name="user_teacher").exists(),
    }
