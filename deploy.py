#!/usr/bin/env python
#

import os
import time

def shell(s):
    print '#',s
    return os.system(s)

shell("appcfg.py -A catchy-link update -V v" + str(int(time.time())) + " ./")
