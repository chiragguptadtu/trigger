# ---trigger---
# name: Clear Cache
# description: Flush the Redis cache for one or more services
# inputs:
#   - name: environment
#     label: Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: services
#     label: Services
#     type: closed
#     options: [api, search, session, all]
#     multi: true
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    env = inputs.get("environment")
    services = inputs.get("services", [])
    print(f"Clearing cache for {services} on {env}")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)
