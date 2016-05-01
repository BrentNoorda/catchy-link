#!/usr/bin/env python
#

import os
import time

def shell(s):
    print '#',s
    return os.system(s)

#shell("appcfg.py -A catchy-link update -V v" + str(int(time.time())) + " ./")

shell("cp app.yaml app.yaml.tmp")
shell("cat secret/live_extensions_app.yaml >> app.yaml")
shell("appcfg.py -A catchy-link update -V v1 ./")
shell("cp app.yaml.tmp app.yaml")
shell("rm app.yaml.tmp")

shell("open http://catchy.link/")
