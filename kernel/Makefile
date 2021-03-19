obj-m += ml_cd.o
ml_cd-y = ml_crash_detector.o
ml_cd-y += state.o
ml_cd-y += communication.o

KDIR := /lib/modules/$(shell uname -r)/build

default:
	$(MAKE) -C $(KDIR) M=$(shell pwd) modules
clean:
	$(MAKE) -C $(KDIR) M=$(shell pwd) clean
