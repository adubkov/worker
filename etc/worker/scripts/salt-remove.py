#!/usr/bin/env python

import sys
import urllib2

from time import time
from hashlib import sha256

def auth_encode(hostname, expires):
    signature = " ".join([str(expires), 'remove', 'secret_key', str(hostname)])
    return sha256(signature).hexdigest()

def delete(hostname, expires, signature):
    url = "http://salt.local:8080/auth?action=remove&host={hostname}&expires={expires}&signature={signature}".format(
            hostname=hostname,
            expires=expires,
            signature=signature)
    req = urllib2.Request(url)
    req.get_method = lambda: 'DELETE'
    res = urllib2.urlopen(req, timeout=10).getcode()
    return res

if len(sys.argv) > 1:
    hostname = sys.argv[1]
    expires = int(time() + 300)
    signature = auth_encode(hostname, expires)
    result = delete(hostname, expires, signature)

    # Output
    if result == 200:
        print '{hostname} removed from salt.'.format(hostname=hostname)
    else:
        sys.exit(1)
else:
    sys.exit(1)
