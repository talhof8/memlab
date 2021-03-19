from django.utils import timezone
from memlab_backend.accounts import models
from memlab_backend.accounts import serializers
from rest_framework import viewsets, status, mixins, permissions, decorators
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response


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
        except KeyError as e:
            raise ValidationError(e)

        new_company = models.Company(name=company_name)
        new_company.save()

        profile = models.Profile.objects.get(user__id=request.user.id)
        profile.company = new_company
        profile.save()

        data = serializer.to_representation(new_company)
        return Response(data, status=status.HTTP_200_OK)

    # todo: support user's company change


class UserViewSet(mixins.ListModelMixin, mixins.CreateModelMixin, viewsets.GenericViewSet):
    serializer_class = serializers.UserSerializer

    permission_classes_by_action = {'create': [permissions.AllowAny]}

    def list(self, request, *args, **kwargs):
        instance = request.user
        serializer = self.get_serializer(instance)
        return Response(serializer.data)

    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            username = validated_data["username"]
            email = validated_data["email"]
            password = validated_data["password"]
            first_name = validated_data["first_name"]
            last_name = validated_data["last_name"]
        except KeyError as e:
            raise ValidationError(e)

        new_user = models.User(username=username, email=email, first_name=first_name, last_name=last_name)
        new_user.set_password(password)
        new_user.save()

        data = serializer.to_representation(new_user)
        return Response(data, status=status.HTTP_200_OK)

    def get_permissions(self):
        try:
            # return permission_classes depending on `action`
            return [permission() for permission in self.permission_classes_by_action[self.action]]
        except KeyError:
            # action is not set return default permission_classes
            return [permission() for permission in self.permission_classes]


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
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            is_company_admin = validated_data["is_company_admin"]
            license_type = validated_data["license_type"]
            license_expires_at = validated_data["license_expires_at"]
        except KeyError as e:
            raise ValidationError(e)

        # Profile already auto-created by user's post-create hook, we only need to update it.
        profile = models.Profile.objects.get(user__id=request.user.id)
        profile.is_company_admin = is_company_admin
        profile.license_type = license_type
        profile.license_modified_at = timezone.now()
        profile.license_expires_at = license_expires_at
        profile.save()

        data = serializer.to_representation(profile)
        return Response(data, status=status.HTTP_200_OK)

    # todo: support profile' updates (e.g: license)
