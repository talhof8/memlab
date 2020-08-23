from django.utils import timezone
from rest_framework import viewsets, mixins, status
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from . import models
from . import serializers


class HostViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.CreateModelMixin, mixins.UpdateModelMixin,
                  viewsets.GenericViewSet):
    queryset = models.Host.objects.all()
    serializer_class = serializers.HostSerializer
    lookup_field = "machine_id"

    # Override create() so that /hosts POST can be used for both creations and updates given a machine id.
    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data, many=True)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            machine_id = validated_data["machine_id"]
            public_ip_address = validated_data["public_ip_address"]
            hostname = validated_data["hostname"]
            last_boot_at = validated_data["last_boot_at"]
            operating_system = validated_data["operating_system"]
            platform = validated_data["platform"]
            platform_family = validated_data["platform_family"]
            platform_version = validated_data["platform_version"]
            kernel_version = validated_data["kernel_version"]
            kernel_architecture = validated_data["kernel_architecture"]
            virtualization_system = validated_data["virtualization_system"]
            virtualization_role = validated_data["virtualization_role"]
        except KeyError as e:  # Just in case...
            raise ValidationError(e)

        # todo: support sending a partial list and omit empty values

        host = models.Host.objects.update_or_create(machine_id=machine_id, defaults={
            "last_keepalive_at": timezone.now(),
            "public_ip_address": public_ip_address,
            "hostname": hostname,
            "last_boot_at": last_boot_at,
            "operating_system": operating_system,
            "platform": platform,
            "platform_family": platform_family,
            "platform_version": platform_version,
            "kernel_version": kernel_version,
            "kernel_architecture": kernel_architecture,
            "virtualization_system": virtualization_system,
            "virtualization_role": virtualization_role
        })

        data = serializer.to_representation(host)
        return Response(data, status=status.HTTP_201_CREATED)

    def get_serializer_context(self):
        return {'request': None}


class ProcessViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.UpdateModelMixin, mixins.CreateModelMixin,
                     viewsets.GenericViewSet):
    queryset = models.Process.objects.all()
    serializer_class = serializers.ProcessSerializer

    def get_serializer_context(self):
        return {'request': None}


class ProcessEventViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.CreateModelMixin,
                          viewsets.GenericViewSet):
    queryset = models.ProcessEvent.objects.all()
    serializer_class = serializers.ProcessEventSerializer

    def get_serializer_context(self):
        return {'request': None}


class ProcessConfigurationViewSet(viewsets.ModelViewSet):
    queryset = models.ProcessConfiguration.objects.all()
    serializer_class = serializers.ProcessConfigurationSerializer

    def get_serializer_context(self):
        return {'request': None}
