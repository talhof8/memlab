# from rest_framework import viewsets
#
# from . import models
# from . import serializers
#
#
# class HostViewSet(viewsets.ModelViewSet):
#     # todo: authenticate by api key
#     queryset = models.Host.objects.all().order_by('pretty_id')
#     serializer_class = serializers.CompanySerializer
#
#
# class UserViewSet(viewsets.ModelViewSet):
#     queryset = models.User.objects.all()
#     serializer_class = serializers.UserSerializer
#
#
# class ProfileViewSet(viewsets.ModelViewSet):
#     queryset = models.Profile.objects.all()
#     serializer_class = serializers.ProfileSerializer
#
#     def get_serializer_context(self):
#         return {'request': None}
