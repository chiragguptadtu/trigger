#!/usr/bin/env python3
# ---trigger---
# name: Provision Environment
# description: Provision a new environment instance in the selected tier and region
# inputs:
#   - name: tier
#     label: Tier
#     type: closed
#     dynamic: true
#     required: true
#   - name: region
#     label: Region
#     type: closed
#     dynamic: true
#     required: true
#   - name: name
#     label: Instance Name
#     type: open
#     required: true
# ---end---

import sys
import json


def get_options(input_name, config):
    """Return dynamic options for closed inputs.

    config keys used:
      DEPLOYMENT_TIERS  — comma-separated list of available tiers
                          (default: dev,staging,production)
      ACTIVE_REGIONS    — comma-separated list of active regions
                          (default: us-east-1,eu-west-1,ap-southeast-1)
    """
    if input_name == "tier":
        raw = config.get("DEPLOYMENT_TIERS", "dev,staging,production")
        return [t.strip() for t in raw.split(",") if t.strip()]
    if input_name == "region":
        raw = config.get("ACTIVE_REGIONS", "us-east-1,eu-west-1,ap-southeast-1")
        return [r.strip() for r in raw.split(",") if r.strip()]
    return []


def run(inputs, config):
    tier = inputs.get("tier")
    region = inputs.get("region")
    name = inputs.get("name")

    print(f"Provisioning '{name}' in tier={tier}, region={region}")
    print("Done.")
    return ""


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--trigger-get-options":
        input_name = sys.argv[2] if len(sys.argv) > 2 else ""
        config = json.loads(sys.argv[3]) if len(sys.argv) > 3 else {}
        print("\n".join(get_options(input_name, config)))
        sys.exit(0)

    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)
