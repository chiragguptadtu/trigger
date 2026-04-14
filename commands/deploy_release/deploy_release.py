# ---trigger---
# name: Deploy Release
# description: Deploy a versioned release to the selected environment with full deployment options
# inputs:
#   - name: environment
#     label: Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: version
#     label: Version Tag
#     type: open
#     required: true
#   - name: services
#     label: Services
#     type: closed
#     options: [api, worker, scheduler, frontend, admin]
#     multi: true
#     required: true
#   - name: region
#     label: Region
#     type: closed
#     options: [us-east-1, us-west-2, eu-west-1, ap-southeast-1]
#     multi: false
#     required: true
#   - name: strategy
#     label: Deploy Strategy
#     type: closed
#     options: [rolling, blue-green, canary]
#     multi: false
#     required: true
#   - name: run_migrations
#     label: Run Migrations
#     type: closed
#     options: [yes, no]
#     multi: false
#     required: true
#   - name: notify_channel
#     label: Notify Slack Channel
#     type: open
#     required: false
#   - name: reason
#     label: Reason
#     type: open
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    print(f"Deploying {inputs.get('version')} to {inputs.get('environment')}")
    print(f"Services: {inputs.get('services')}")
    print(f"Region: {inputs.get('region')}, Strategy: {inputs.get('strategy')}")
    print(f"Migrations: {inputs.get('run_migrations')}, Reason: {inputs.get('reason')}")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)
