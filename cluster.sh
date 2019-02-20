#!/bin/bash

vm_list() {
    local node_name_prefix="l23_${USER}"
    if [[ "${VAGRANT_L23_NAME_SUFFIX}" != "" ]] ; then
        node_name_prefix="${node_name_prefix}-${VAGRANT_L23_NAME_SUFFIX}"
    fi

    local VMs=$(virsh list --all | grep "${node_name_prefix}" | awk '{print $2}' | sort)

    if [[ "${VMs}" == "" ]] ; then
        echo "No VMs created for you env"
        exit 1
    fi

    echo $VMs
}

COMMAND=$1
if [[ "${COMMAND}" == "create" ]] ; then
    echo "Cluster will be created."
    vagrant up
    rc=$?
    if [[ "${rc}" == "0" ]] && [[ -z "${VAGRANT_L23_NO_PROVISION}" ]]; then
      echo "Provisioning started:"
      ANSIBLE_FORCE_COLOR=true ANSIBLE_HOST_KEY_CHECKING=false ANSIBLE_SSH_ARGS='-o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes -o ControlMaster=auto -o ControlPersist=60s' ansible-playbook --timeout=30 --inventory-file=.vagrant/provisioners/ansible/inventory --become -v func_tests/playbooks/cluster.yaml
      rc=$?
      echo "Provisioning done, rc=${rc}"
    fi
    exit $rc
elif [[ "${COMMAND}" == "destroy" ]] ; then
    vagrant destroy -f
    exit $?
elif [[ "${COMMAND}" == "ssh" ]] ; then
    vagrant ssh $2 -c 'sudo -i'
    exit $?
elif [[ "${COMMAND}" =~ ^snapshot-.* ]] ; then
    echo "."
else
    echo "[err] unsupported command."
    exit 1
fi

### work with snapshots
SNAPSHOT_NAME=$2
if [[ "${SNAPSHOT_NAME}" == "" && "${COMMAND}" != "snapshot-list" ]] ; then
    echo "No snapshot name given"
    exit 1
fi
VMs=$(vm_list)

if [[ "${COMMAND}" == "snapshot-create" ]] ; then
    for i in $VMs ; do virsh suspend $i ; done
    for i in $VMs ; do virsh snapshot-create-as $i $SNAPSHOT_NAME ; done
    for i in $VMs ; do virsh resume $i ; done
    exit $?
elif [[ "${COMMAND}" == "snapshot-delete" ]] ; then
    for i in $VMs ; do virsh snapshot-delete $i --snapshotname=$SNAPSHOT_NAME ; done
    exit $?
elif [[ "${COMMAND}" == "snapshot-list" ]] ; then
    SS=""
    for i in $VMs ; do SS="${SS} $(virsh snapshot-list ${i} | grep -v -e ' Name ' -e '---------' | awk '{print $1}')"; done
    echo $SS | perl -pe 's/\s+/\n/g' | sort | uniq
    exit $?
elif [[ "${COMMAND}" == "snapshot-revert" ]] ; then
    for i in $VMs ; do virsh suspend $i ; done
    for i in $VMs ; do virsh snapshot-revert $i $SNAPSHOT_NAME ; done
    for i in $VMs ; do virsh resume $i ; done
    exit $?
else
    echo "[err] unsupported command."
    exit 1
fi
