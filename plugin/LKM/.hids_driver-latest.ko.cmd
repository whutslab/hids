cmd_/etc/hids/plugin/LKM/hids_driver-latest.ko := ld -r -m elf_x86_64 -z max-page-size=0x200000 -T ./scripts/module-common.lds --build-id  -o /etc/hids/plugin/LKM/hids_driver-latest.ko /etc/hids/plugin/LKM/hids_driver-latest.o /etc/hids/plugin/LKM/hids_driver-latest.mod.o ;  true