# Generated by Django 3.1 on 2020-08-19 15:32

from django.db import migrations, models
import django.utils.timezone


class Migration(migrations.Migration):

    dependencies = [
        ('hosts', '0003_auto_20200819_1532'),
    ]

    operations = [
        migrations.AlterField(
            model_name='host',
            name='last_activity_at',
            field=models.DateTimeField(default=django.utils.timezone.now),
        ),
    ]
