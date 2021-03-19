def add_user_to_validated_data(request, validated_data):
    validated_data["user"] = request.user
