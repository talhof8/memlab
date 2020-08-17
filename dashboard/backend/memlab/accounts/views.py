from rest_framework import viewsets, status
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from . import models
from . import serializers


class CompanyViewSet(viewsets.ModelViewSet):
    serializer_class = serializers.CompanySerializer

    def get_queryset(self):
        return models.Company.objects.filter(profile__user__id=self.request.user)

    def list(self, request, *args, **kwargs):
        instance = models.Company.objects.get(profile__user__id=self.request.user)
        serializer = self.get_serializer(instance)
        return Response(serializer.data)

    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            company_name = validated_data["name"]
        except KeyError as e:  # Just in case...
            raise ValidationError(e)

        new_company = models.Company(name=company_name)
        new_company.save()

        profile = models.Profile.objects.get(user__id=request.user.id)
        profile.company = new_company
        profile.save()

        data = serializer.to_representation(new_company)
        headers = self.get_success_headers(data)
        return Response(data, status=status.HTTP_201_CREATED, headers=headers)


class UserViewSet(viewsets.ModelViewSet):
    serializer_class = serializers.UserSerializer

    def get_queryset(self):
        return models.User.objects.get(id=self.request.user.id)

    def list(self, request, *args, **kwargs):
        instance = request.user
        serializer = self.get_serializer(instance)
        return Response(serializer.data)


class ProfileViewSet(viewsets.ModelViewSet):
    serializer_class = serializers.ProfileSerializer

    def get_queryset(self):
        return models.Profile.objects.filter(user__id=self.request.user.id)

    def list(self, request, *args, **kwargs):
        instance = models.Profile.objects.get(user__id=self.request.user.id)
        serializer = self.get_serializer(instance)
        return Response(serializer.data)

    def get_serializer_context(self):
        return {'request': None}  # So url is relative and not absolute
