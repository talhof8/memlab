from memlab_backend.accounts import models
from rest_framework import serializers, reverse


class CompanySerializer(serializers.ModelSerializer):
    class Meta:
        model = models.Company
        fields = ["name", "pretty_id", "created_at"]
        read_only_fields = ["pretty_id", "created_at"]


class UserSerializer(serializers.ModelSerializer):
    class Meta:
        model = models.User
        fields = ["username", "email", "first_name", "last_name", "password"]
        extra_kwargs = {
            "password": {"write_only": True}  # Note: Super important!!! so that password hashes do not leak out.
        }


class ProfileSerializer(serializers.ModelSerializer):
    user = serializers.SerializerMethodField("get_user")
    company = serializers.SerializerMethodField("get_company")

    class Meta:
        model = models.Profile
        fields = ["user", "company", "created_at", "is_company_admin", "license_type", "license_modified_at",
                  "license_expires_at"]
        read_only_fields = ["user", "company", "license_modified_at", "created_at"]

    def get_user(self, obj):
        return reverse.reverse("user-list")

    def get_company(self, obj):
        return reverse.reverse("company-list")
