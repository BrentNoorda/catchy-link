#!/usr/bin/env python
#

import os
import time
import thread

def shell(s):
    print '#',s
    return os.system(s)

def open_browser_windows():
    shell("open http://localhost:8000/ &")
    shell("open http://localhost:8080/ &")


retCode = shell("goapp build")
if retCode == 0:
    open_browser_windows()
    shell("goapp serve ./")
