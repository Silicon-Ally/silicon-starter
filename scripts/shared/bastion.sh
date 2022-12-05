#!/bin/bash
set -euo pipefail

declare BASTION_INSTANCE_NAME
declare BASTION_PROJECT_NUMBER
declare BASTION_GCP_PROJECT
declare BASTION_GCP_ZONE

CONTROL_SOCKET_DIR="$(mktemp -d -t bastion-XXXXXXXXX)"
trap 'rm -rf -- "$CONTROL_SOCKET_DIR"' EXIT
CONTROL_SOCKET_FILE="${CONTROL_SOCKET_DIR}/sql-ctrl-socket"

function join {
  local joined="$(printf ",%s" "$@")"
  printf '%s' "${joined:1}"
}

function start_tunnel {

  declare -a SCOPES=(
    "https://www.googleapis.com/auth/devstorage.read_only"
    "https://www.googleapis.com/auth/logging.write"
    "https://www.googleapis.com/auth/monitoring.write"
    "https://www.googleapis.com/auth/servicecontrol"
    "https://www.googleapis.com/auth/service.management.readonly"
    "https://www.googleapis.com/auth/trace.append"
  )

  if ! [ -x "$(command -v gcloud)" ]; then
    echo 'Error: gcloud is not installed.' >&2
    exit 1
  fi

  declare PRIVATE_POSTGRES_ADDR
  case "$1" in
    dev)
      BASTION_INSTANCE_NAME="bastion-${USER}"
      BASTION_PROJECT_NUMBER="<dev project number>"
      BASTION_GCP_PROJECT="<dev project ID>"
      BASTION_GCP_ZONE="<zone in dev project region>"
      PRIVATE_POSTGRES_ADDR='10.x.y.z'
      ;;
    *)
      echo "Unknown environment ${1}"
      exit 1
      ;;
  esac

  CREATE_DISK=(
    "auto-delete=yes"
    "boot=yes"
    "device-name=${BASTION_INSTANCE_NAME}"
    "image=projects/debian-cloud/global/images/debian-10-buster-v20220822"
    "mode=rw"
    "size=10"
    "type=projects/${BASTION_GCP_PROJECT}/zones/${BASTION_GCP_ZONE}/diskTypes/pd-balanced"
  )

  gcloud compute instances describe "$BASTION_INSTANCE_NAME" \
    --project="$BASTION_GCP_PROJECT" \
    --zone="$BASTION_GCP_ZONE" >/dev/null 2>&1 && INSTANCE_EXIT_CODE="$?" || INSTANCE_EXIT_CODE="$?"
  if [ $INSTANCE_EXIT_CODE -ne 0 ]; then
    echo "Creating bastion host..."
    gcloud compute instances create "$BASTION_INSTANCE_NAME" \
      --project="$BASTION_GCP_PROJECT" \
      --zone="$BASTION_GCP_ZONE" \
      --machine-type=e2-micro \
      --network-interface=subnet=default,no-address \
      --maintenance-policy=MIGRATE \
      --service-account="deployer@${BASTION_GCP_PROJECT}.iam.gserviceaccount.com" \
      --scopes="$(join ${SCOPES[@]})" \
      --tags=bastion \
      --create-disk="$(join ${CREATE_DISK[@]})" \
      --no-shielded-secure-boot \
      --shielded-vtpm \
      --shielded-integrity-monitoring \
      --reservation-affinity=any
  else
    echo "Using existing bastion host"
  fi

  SSH_EXIT_CODE="1"
  while [ $SSH_EXIT_CODE -ne 0 ]; do
    echo "Starting SSH tunnel to instance..."
    gcloud compute ssh "$BASTION_INSTANCE_NAME" \
      --project="$BASTION_GCP_PROJECT" \
      --zone="$BASTION_GCP_ZONE" \
      --tunnel-through-iap -- \
        -fnNT \
        -L "localhost:5433:${PRIVATE_POSTGRES_ADDR}:5432" \
        -M -S "$CONTROL_SOCKET_FILE" >/dev/null 2>&1 && SSH_EXIT_CODE="$?" || SSH_EXIT_CODE="$?"
    sleep 3
  done
  echo "Started SSH tunnel to instance..."

  trap 'stop_tunnel' EXIT
}

stop_tunnel() {
  echo "Closing tunnel..."
  ssh -S "$CONTROL_SOCKET_FILE" -O exit unused

  rm -rf -- "$CONTROL_SOCKET_DIR"

  echo "Shutting down bastion host..."
  gcloud compute instances delete "$BASTION_INSTANCE_NAME" \
    --project="$BASTION_GCP_PROJECT" \
    --zone="$BASTION_GCP_ZONE" \
    --delete-disks=all \
    --quiet
}
