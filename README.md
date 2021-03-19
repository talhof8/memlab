# Memlab - A new way to do production debugging

## Minimum Requirements
### Agent
- ProcDump (https://github.com/microsoft/ProcDump-for-Linux):
    - Kernel Version: 3.5+
    - Minimum OS:
        Red Hat Enterprise Linux / CentOS 7
        Fedora 29
        Ubuntu 16.04 LTS
    - gdb >= 7.6.1
    - zlib (build-time only)

### Kernel Module
- Kernel version: 3.17+ (determined by use of linux/rhashtable.h).
- ftrace enabled

### Backend
- Python 2.7+ (Developed & Tested on Python 3.8)