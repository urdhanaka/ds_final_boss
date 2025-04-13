#!/bin/bash

BRIDGE_NAME=%s
ETHERNET_INTERFACE=%s

# template for creating bridge network on node
# because this file is a template, some golang format (%s)
# is used
# DO NOT USE AS IT IS, THIS FILE ONLY USED WITH GOLANG FORMATF
nmcli connection add type bridge con-name $BRIDGE_NAME ifname $BRIDGE_NAME
nmcli connection add type bridge-slave ifname $ETHERNET_INTERFACE master $BRIDGE_NAME

# turn off current ethernet connection
# and then start bridge network
nmcli connection down $ETHERNET_INTERFACE
nmcli connection up $BRIDGE_NAME
