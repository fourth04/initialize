#!/bin/bash

DPDK_NICCONF_FILE="/etc/sdpi/dpdk_nic_config"

pci=$(dpdk-devbind.py --status | grep 'drv=igb_uio' | awk '{print $1}')
for i in $pci;
do
        dpdk-devbind.py -b igb $i
done

rm -f $DPDK_NICCONF_FILE
