from rest_framework import serializers, reverse

from . import models


class CompanySerializer(serializers.ModelSerializer):
    class Meta:
        model = models.Company
        fields = "__all__"
        read_only_fields = ["id", "pretty_id", "created_at"]


class UserSerializer(serializers.ModelSerializer):
    class Meta:
        model = models.User
        fields = ["username", "email", "first_name", "last_name"]
        read_only_fields = ["id", "username", "is_staff", "is_active", "date_joined"]


class ProfileSerializer(serializers.ModelSerializer):
    user = serializers.SerializerMethodField("get_user")
    company = serializers.SerializerMethodField("get_company")
    license_type = serializers.CharField(source="get_license_type_display")

    class Meta:
        model = models.Profile
        fields = ["user", "company", "created_at", "is_company_admin", "license_type", "license_modified_at",
                  "license_expires_at"]
        read_only_fields = ["user", "company", "created_at"]

    def get_user(self, obj):
        return reverse.reverse('user-list')

    def get_company(self, obj):
        return reverse.reverse('company-list')
