from django.contrib import admin

from memlab_backend.hosts import models


class HostAdmin(admin.ModelAdmin):
    readonly_fields = ("first_seen", "last_probe_at",)


class ProcessAdmin(admin.ModelAdmin):
    readonly_fields = ("last_seen_at",)


class ProcessEventAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at",)


class DetectionConfigAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at", "modified_at")


admin.site.register(models.Host, HostAdmin)
admin.site.register(models.Process, ProcessAdmin)
admin.site.register(models.ProcessEvent, ProcessEventAdmin)
admin.site.register(models.DetectionConfig, DetectionConfigAdmin)
