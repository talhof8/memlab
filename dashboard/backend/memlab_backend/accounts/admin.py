from django.contrib import admin

from memlab_backend.accounts import models


class CompanyAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at",)


class ProfileAdmin(admin.ModelAdmin):
    readonly_fields = ("created_at",)


admin.site.register(models.Company, CompanyAdmin)
admin.site.register(models.User)
admin.site.register(models.Profile, ProfileAdmin)
