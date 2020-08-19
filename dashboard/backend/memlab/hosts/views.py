from rest_framework import viewsets

from . import models
from . import serializers


class HostViewSet(viewsets.ModelViewSet):
    queryset = models.Host.objects.all()
    serializer_class = serializers.HostSerializer

    def get_serializer_context(self):
        return {'request': None}


class ProcessViewSet(viewsets.ModelViewSet):
    queryset = models.Process.objects.all()
    serializer_class = serializers.ProcessSerializer

    def get_serializer_context(self):
        return {'request': None}


class ProcessEventViewSet(viewsets.ModelViewSet):
    queryset = models.ProcessEvent.objects.all()
    serializer_class = serializers.ProcessEventSerializer

    def get_serializer_context(self):
        return {'request': None}


class ProcessConfigurationViewSet(viewsets.ModelViewSet):
    queryset = models.ProcessConfiguration.objects.all()
    serializer_class = serializers.ProcessConfigurationSerializer

    def get_serializer_context(self):
        return {'request': None}
