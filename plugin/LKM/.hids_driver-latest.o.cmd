cmd_/etc/hids/plugin/LKM/hids_driver-latest.o := ld -m elf_x86_64 -z max-page-size=0x200000  -r -T /etc/hids/plugin/LKM/kprobe.lds  -r -o /etc/hids/plugin/LKM/hids_driver-latest.o /etc/hids/plugin/LKM/src/init.o /etc/hids/plugin/LKM/src/kprobe.o /etc/hids/plugin/LKM/src/trace.o /etc/hids/plugin/LKM/src/smith_hook.o /etc/hids/plugin/LKM/src/anti_rootkit.o /etc/hids/plugin/LKM/src/filter.o /etc/hids/plugin/LKM/src/util.o 