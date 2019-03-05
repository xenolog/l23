# -*- mode: ruby -*-
# vi: set ft=ruby :

require "yaml"

class ::Hash
    def deep_merge(second)
        merger = proc { |key, v1, v2| Hash === v1 && Hash === v2 ? v1.merge(v2, &merger) : Array === v1 && Array === v2 ? v1 | v2 : [:undefined, nil, :nil].include?(v2) ? v1 : v2 }
        self.merge(second.to_h, &merger)
    end
end

Vagrant.require_version ">= 2.1.5"

# workaround for https://app.vagrantup.com/boxes/ usage
# Vagrant::DEFAULT_SERVER_URL.replace('https://vagrantcloud.com')

$pool = ENV["VAGRANT_L23_POOL"] || "10.251.0.0/16"
$prefix = $pool.gsub(/\.\d+\.\d+\/\d+$/, "")
vagrant_cidr = $prefix + ".0.0/24"

ENV["VAGRANT_DEFAULT_PROVIDER"] = "libvirt"

# Boxes with libvirt provider support:
box = ENV["VAGRANT_L23_BOX"] || "generic/ubuntu1604"


vm_memory = (ENV["VAGRANT_L23_NODE_MEMORY"] || "1024").to_i
vm_cpus = (ENV["VAGRANT_L23_NODE_CPUS"] || "1").to_i
master_memory = (ENV["VAGRANT_L23_MASTER_MEMORY"] || "1024").to_i
master_cpus = (ENV["VAGRANT_L23_MASTER_CPUS"] || "1").to_i

num_networks = (ENV["VAGRANT_L23_NETWORKS"] || "3").to_i
num_slaves = (ENV["VAGRANT_L23_SLAVES"] || "1").to_i

user = ENV["USER"]
node_name_prefix = "#{user}"
node_name_suffix = ENV["VAGRANT_L23_NAME_SUFFIX"] || ""
node_name_prefix+= "-#{node_name_suffix}" if node_name_suffix != ""


print("### ENV:\n" +
      "    master node MEM: #{master_memory}\n"+
      "    master node CPU: #{master_cpus}\n"+
      "          nodes MEM: #{vm_memory}\n"+
      "          nodes CPU: #{vm_cpus}\n"+
      "     control subnet: #{vagrant_cidr}\n"
)


# (1..num_racks).each do |rack_no|
#   nodes_per_rack << (ENV["VAGRANT_L23_RACK#{rack_no}_NODES"] || "2").to_i
#   rack_subnets << (ENV["VAGRANT_L23_RACK#{rack_no}_CIDR"] || prefix.to_s + ".#{rack_no}.0/24")
#   cp_nodes = (ENV["VAGRANT_L23_RACK#{rack_no}_CP_NODES"] || "1").split(',').select{|x| x.to_i > 0}
#   cp_nodes = [1] if cp_nodes.size <1
#   cp_nodes_for_rack << cp_nodes.map{|x| x.to_i}
#   print("### RACK #{rack_no}:\n" +
#         "    nodes: #{nodes_per_rack[rack_no]}  (CP: #{cp_nodes_for_rack[rack_no].join(',')})\n"+
#         "   subnet: #{rack_subnets[rack_no]}\n\n")
# end


# Create SSH keys for future lab
# system "bash scripts/ssh-keygen.sh"

# Create nodes list for future kargo deployment
# nodes=[]
# (1..num_racks).each do |rack_no|
#   (1..nodes_per_rack[rack_no]).each do |node_no|
#     nodes << rack_subnets[rack_no].split(".")[0..2].join(".")+".#{node_no}"
#   end
# end
# File.open("tmp/nodes", "w") do |file|
#   file.write(nodes.join("\n"))
#   file.write("\n")
# end

# prepare ansible deployment facts for master and slave nodes
# This hash should be assembled before run any provisioners for prevent
# parallel provisioning race conditions
master_node_name = "%s-000" % [node_name_prefix]
# master_node_ipaddr = public_subnets[0].split(".")[0..2].join(".")+".254"

slave_nodes_num = (ENV["VAGRANT_L23_SLAVE_NODES"] || "1").to_i
nodes = []
ansible_host_vars = {}
(1..slave_nodes_num).each do |node_no|
    slave_name = "%s-%02d" % [node_name_prefix, node_no]
    nodes << slave_name
    ansible_host_vars[slave_name] = {
        "ansible_python_interpreter" => "/usr/bin/python3",
        "node_name"                  => slave_name,
        "master_node_name"           => master_node_name,
        # "master_node_ipaddr"         => master_node_ipaddr,
        "node_no"                    => "'%03d'" % node_no,
    }
end
ansible_host_vars[master_node_name] = {
  "ansible_python_interpreter" => "/usr/bin/python3",
  "node_name"                  => "#{master_node_name}",
  "master_node_name"           => "#{master_node_name}",
#   "master_node_ipaddr"         => "#{master_node_ipaddr}",
}


def common_config(node, config, nn, np, node_no)
    # isolated networks for tests
    (1..nn).each do |net_no|
        node_ip = $prefix + "." + net_no.to_s + "." + node_no.to_s
        node.vm.network(:private_network,
          :ip => node_ip,
          :libvirt__host_ip => node_ip.split(".")[0..2].join(".")+".254",
          :model_type => "e1000",
          :libvirt__network_name => "l23_#{np}_test_%02d" % [net_no],
          :libvirt__dhcp_enabled => false,
          :libvirt__forward_mode => "none"
        )
    end
    config.vm.synced_folder ".", "/vagrant", disabled: true
    config.vm.synced_folder ".", "/root/go/l23", create: true
end

# Create the lab
Provisioner = nil
Vagrant.configure("2") do |config|
  config.ssh.insert_key = false
  config.vm.box = box

  # This fake ansible provisioner required for creating inventory for
  # true ansible privisioner, which should be run outside Vagrant
  # due Vagrant used featured method to run Ansible with unwanted features.
  config.vm.provision :ansible, preserve_order: true do |a|
    a.become = true  # it's a sudo !!!
    a.playbook = "func_tests/playbooks/fake.yaml"
    #a.extra_vars = {}
    a.host_vars = ansible_host_vars
    a.verbose = true
    a.limit = "all"  # fake provisioner will be run
    a.groups = {
      "nodes"   => nodes,
      "masters" => [master_node_name],
    #   "masters:vars" => {
    #     "virt_racks" => racks,
    #   },
    }
  end

  # configure Master VM
  config.vm.define "#{master_node_name}", primary: true do |master_node|
    master_node.vm.hostname = "#{master_node_name}"
    # Libvirt provider settings
    master_node.vm.provider(:libvirt) do |domain|
      domain.uri = "qemu+unix:///system"
      domain.memory = master_memory
      domain.cpus = master_cpus
      domain.driver = "kvm"
      domain.host = "localhost"
      domain.connect_via_ssh = false
      domain.username = user
      domain.storage_pool_name = "default"
      domain.nic_model_type = "e1000"
      domain.management_network_name = "l23_#{node_name_prefix}_vagrant"
      domain.management_network_address = "#{vagrant_cidr}"
      domain.nested = true
      domain.cpu_mode = "host-passthrough"
      domain.volume_cache = "unsafe"
      domain.disk_bus = "virtio"
      # DISABLED: switched to new box which has 100G / partition
      #domain.storage :file, :type => "qcow2", :bus => "virtio", :size => "20G", :device => "vdb"
    end
    common_config(master_node, config, num_networks, node_name_prefix, 0)
  end

  # Slave VMs
  (1..num_slaves).each do |node_no|
      slave_name = "%s-%02d" % [node_name_prefix, node_no]
      config.vm.define "#{slave_name}" do |slave_node|
        slave_node.vm.hostname = "#{slave_name}"
        # Libvirt provider settings
        slave_node.vm.provider :libvirt do |domain|
          domain.uri = "qemu+unix:///system"
          domain.memory = vm_memory
          domain.cpus = vm_cpus
          domain.driver = "kvm"
          domain.host = "localhost"
          domain.connect_via_ssh = false
          domain.username = user
          domain.storage_pool_name = "default"
          domain.nic_model_type = "e1000"
          domain.management_network_name = "l23_#{node_name_prefix}_vagrant"
          domain.management_network_address = "#{vagrant_cidr}"
          domain.nested = true
          domain.cpu_mode = "host-passthrough"
          domain.volume_cache = "unsafe"
          domain.disk_bus = "virtio"
          # DISABLED: switched to new box which has 100G / partition
          #domain.storage :file, :type => "qcow2", :bus => "virtio", :size => "20G", :device => "vdb"
        end
        common_config(slave_node, config, num_networks, node_name_prefix, node_no)
    end
  end
end
