#!/bin/bash

DPDK_NICCONF_FILE="/etc/sdpi/dpdk_nic_config"
PROG_CONF_FILE="/etc/sdpi/sdpi.ini"

INI_SECTION_DPI="dpi"
INI_SECTION_DNS="dns"
INI_KEY_DPI_IN_NIC="in_nic"
INI_KEY_DPI_OUT_NIC="out_nic"
INI_KEY_DNS_IN_NIC="in_nic"
INI_KEY_DNS_OUT_NIC="out_nic"

. /usr/bin/ini-op.sh

declare -a NIC_INF
NIC_INF_INDEX=0

nic_inf_insert()
{
        local nic

        for(( idx=0;idx<NIC_INF_INDEX;idx++ ))
        do
                nic=${NIC_INF[idx]}
                if [ "$nic" = "$1" ]
                then
                        return;
                fi
        done

        NIC_INF[NIC_INF_INDEX]="$1"
        let NIC_INF_INDEX++
}

write_nicinfo_to_file()
{
        local srv=$1
        local action=$2
        local name=$3
        local mac=$(ifconfig "$name" | grep ether | awk -F ' ' '{print $2}')

        echo "$srv|$action|$name|$mac" >> $DPDK_NICCONF_FILE
}

set_dpi_nic()
{
        local in_nic=$(readIniKeyValue "$PROG_CONF_FILE" "$INI_SECTION_DPI" "$INI_KEY_DPI_IN_NIC")
        for i in `echo "$in_nic" | awk -F "," '{for (i=1;i<=NF;i++) {print $i}}'`
        do
                write_nicinfo_to_file "dpi" "in_nic" "$i"
                nic_inf_insert "$i"
        done

        local out_nic=$(readIniKeyValue "$PROG_CONF_FILE" "$INI_SECTION_DPI" "$INI_KEY_DPI_OUT_NIC")
        for i in `echo "$out_nic" | awk -F "," '{for (i=1;i<=NF;i++) {print $i}}'`
        do
                write_nicinfo_to_file "dpi" "out_nic" "$i"
                nic_inf_insert "$i"
        done
}

set_dns_nic()
{
        local in_nic=$(readIniKeyValue "$PROG_CONF_FILE" "$INI_SECTION_DNS" "$INI_KEY_DNS_IN_NIC")
        for i in `echo "$in_nic" | awk -F "," '{for (i=1;i<=NF;i++) {print $i}}'`
        do
                write_nicinfo_to_file "dns" "in_nic" "$i"
                nic_inf_insert "$i"
        done

        local out_nic=$(readIniKeyValue "$PROG_CONF_FILE" "$INI_SECTION_DNS" "$INI_KEY_DNS_OUT_NIC")
        for i in `echo "$out_nic" | awk -F "," '{for (i=1;i<=NF;i++) {print $i}}'`
        do
                write_nicinfo_to_file "dns" "out_nic" "$i"
                nic_inf_insert "$i"
        done
}

bind_nic_to_igb_uio()
{
        local nic

        for(( idx=0;idx<NIC_INF_INDEX;idx++ ))
        do
                nic=${NIC_INF[idx]}

                ifconfig "$nic" down
                dpdk-devbind.py -b igb_uio "$nic"
        done
}

set_dpi_nic
set_dns_nic
bind_nic_to_igb_uio
