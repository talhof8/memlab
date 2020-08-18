from rest_framework import viewsets, status, mixins, permissions, decorators
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from . import models
from . import serializers


class CompanyViewSet(mixins.ListModelMixin, mixins.CreateModelMixin, viewsets.GenericViewSet):
    serializer_class = serializers.CompanySerializer

    def list(self, request, *args, **kwargs):
        instance = models.Company.objects.get(profile__user__id=self.request.user.id)
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


class UserViewSet(mixins.ListModelMixin, viewsets.GenericViewSet):
    serializer_class = serializers.UserSerializer

    def list(self, request, *args, **kwargs):
        instance = request.user
        serializer = self.get_serializer(instance)
        return Response(serializer.data)


class ProfileViewSet(viewsets.GenericViewSet):
    serializer_class = serializers.ProfileSerializer

    @decorators.action(detail=False, methods=['get', 'put', 'delete'], url_path='profile')
    def current_profile(self, request, *args, **kwargs):
        instance = models.Profile.objects.get(user__id=self.request.user.id)

        if request.method == 'GET':  # Retrieve profile
            serializer = self.get_serializer(instance)
            return Response(serializer.data)
        elif request.method == 'PUT':  # Update profile
            serializer = self.get_serializer(instance, data=request.data)
            if serializer.is_valid():
                serializer.save()
                return Response(serializer.data)
            return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)
        elif request.method == 'DELETE':  # Delete profile
            instance.delete()
            return Response(status=status.HTTP_204_NO_CONTENT)

        return Response(status=status.HTTP_400_BAD_REQUEST)

    @decorators.permission_classes([permissions.AllowAny])
    @current_profile.mapping.post
    def new_profile(self, request, *args, **kwargs):
        return Response(status=status.HTTP_201_CREATED)
