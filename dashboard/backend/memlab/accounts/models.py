import hashlib
import uuid

from django.contrib.auth.models import AbstractUser
from django.db import models
from django.db.models.signals import post_save, pre_save
from django.dispatch import receiver
from django.utils import timezone
from rest_framework.authtoken.models import Token


class Company(models.Model):
    class Meta:
        verbose_name_plural = "companies"  # Otherwise, it is set to "companys"

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    pretty_id = models.CharField(max_length=15, unique=True)
    name = models.CharField(max_length=150, blank=False, null=False)
    created_at = models.DateTimeField(auto_now_add=True)


@receiver(pre_save, sender=Company)
def pre_create_company(sender, instance, *args, **kwargs):
    if not instance.pretty_id:
        company_name = instance.name.lower()  # Note: company name is being lowered.
        hash_salt = "{id}+{name}".format(id=instance.id, name=company_name)
        instance.pretty_id = hashlib.md5(bytearray(hash_salt, 'utf-8')).hexdigest()


class User(AbstractUser):
    """auth/login-related fields"""
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)


@receiver(post_save, sender=User)
def post_create_user(sender, instance, created, **kwargs):
    if created:
        Token.objects.create(user=instance)
        Profile.objects.create(user=instance)


@receiver(post_save, sender=User)
def post_save_user(sender, instance, **kwargs):
    instance.profile.save()


class Profile(models.Model):
    LICENSE_FREE_USER = 'A'
    LICENSE_PREMIUM_USER = 'B'
    LICENSES = [
        (LICENSE_FREE_USER, 'Free'),
        (LICENSE_PREMIUM_USER, 'Premium'),
    ]

    API_KEN_LENGTH = 50

    """non-auth-related/cosmetic fields"""
    user = models.OneToOneField(User, on_delete=models.CASCADE)
    created_at = models.DateTimeField(auto_now_add=True)
    company = models.ForeignKey(Company, on_delete=models.CASCADE, null=True, blank=True)
    is_company_admin = models.BooleanField(default=False)
    license_type = models.CharField(max_length=1, choices=LICENSES, default=LICENSE_FREE_USER)
    license_modified_at = models.DateTimeField(default=timezone.now)
    license_expires_at = models.DateTimeField(null=True, blank=True)
