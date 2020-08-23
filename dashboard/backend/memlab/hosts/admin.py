from django.contrib import admin

from . import models


class HostAdmin(admin.ModelAdmin):
    readonly_fields = ("first_seen",)


class ProcessEventAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at",)


class ProcessConfigurationAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at", "modified_at")


admin.site.register(models.Host, HostAdmin)
admin.site.register(models.Process)
admin.site.register(models.ProcessEvent, ProcessEventAdmin)
admin.site.register(models.ProcessConfiguration, ProcessConfigurationAdmin)
