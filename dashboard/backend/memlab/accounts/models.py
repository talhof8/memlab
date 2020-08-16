import hashlib
import secrets
import uuid

from django.contrib.auth.models import AbstractUser
from django.db import models
from django.db.models.signals import post_save, pre_save
from django.dispatch import receiver
from django.utils import timezone


class Company(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    pretty_id = models.CharField(max_length=15, unique=True)
    name = models.CharField(max_length=150, blank=False, null=False)


@receiver(pre_save, sender=Company)
def pre_create_company(sender, instance, *args, **kwargs):
    if not instance.pretty_id:
        hash_salt = "{id}+{name}".format(id=instance.id, name=instance.name)
        instance.pretty_id = hashlib.md5(bytearray(hash_salt)).hexdigest()


class User(AbstractUser):
    """auth/login-related fields"""
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)


@receiver(post_save, sender=User)
def post_create_user(sender, instance, created, **kwargs):
    if created:
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
    company = models.ForeignKey(Company, on_delete=models.CASCADE, null=True, blank=True)
    is_company_admin = models.BooleanField(default=False)
    api_key = models.CharField(max_length=API_KEN_LENGTH, null=False, blank=False, unique=True)
    license_type = models.CharField(max_length=1, choices=LICENSES, default=LICENSE_FREE_USER)
    license_modified_at = models.DateTimeField(default=timezone.now)
    license_expires_at = models.DateTimeField(null=True, blank=True)


@receiver(pre_save, sender=Profile)
def pre_create_profile(sender, instance, *args, **kwargs):
    if not instance.api_key:
        instance.api_key = secrets.token_urlsafe(Profile.API_KEN_LENGTH)
