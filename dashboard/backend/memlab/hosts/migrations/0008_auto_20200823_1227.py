# Generated by Django 3.1 on 2020-08-23 12:27

from django.db import migrations


class Migration(migrations.Migration):

    dependencies = [
        ('hosts', '0007_auto_20200823_1128'),
    ]

    operations = [
        migrations.RenameField(
            model_name='host',
            old_name='last_reboot',
            new_name='last_boot_at',
        ),
    ]
