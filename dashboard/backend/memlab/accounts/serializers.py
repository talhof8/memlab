from rest_framework import serializers

from . import models


class CompanySerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.Company
        fields = "__all__"
        read_only_fields = ["id", "pretty_id", "created_at"]


class UserSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.User
        fields = ["username", "email", "first_name", "last_name"]
        read_only_fields = ["id", "username", "is_staff", "is_active", "date_joined"]


class ProfileSerializer(serializers.HyperlinkedModelSerializer):
    user = UserSerializer
    company = CompanySerializer
    license_type = serializers.CharField(source="get_license_type_display")

    class Meta:
        model = models.Profile
        fields = ["user", "company", "created_at", "is_company_admin", "license_type", "license_modified_at",
                  "license_expires_at"]
        read_only_fields = ["id", "user", "company", "created_at"]
