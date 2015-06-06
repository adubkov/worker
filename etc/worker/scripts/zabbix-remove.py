#!/usr/bin/env python

import sys
from zabbix.api import ZabbixAPI

if len(sys.argv) > 1:
    hostname = sys.argv[1]

z = ZabbixAPI(
        url='https://zabbix.local',
        user='admin',
        password='zabbix')

# Get zabbix host id
hostId = z.get_id('host', hostname)

# Remove host by id
result = z.do_request('host.delete', [hostId]).get('result', {}).get('hostids',[None])[0]

# Output
if result:
    print '{hostname} removed from zabbix.'.format(hostname=hostname)
else:
    sys.exit(1)
