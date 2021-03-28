package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

var xml = `
<domain type='kvm' id='5'>
  <name>instance-1</name>
  <uuid>8902ecd3-a6e9-4e59-83d5-4c05a325f67a</uuid>
  <metadata>
    <libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
      <libosinfo:os id="http://ubuntu.com/ubuntu/16.04"/>
    </libosinfo:libosinfo>
  </metadata>
  <memory unit='KiB'>1048576</memory>
  <currentMemory unit='KiB'>1048576</currentMemory>
  <vcpu placement='static'>1</vcpu>
  <resource>
    <partition>/machine</partition>
  </resource>
  <os>
    <type arch='x86_64' machine='pc-i440fx-focal'>hvm</type>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
  </features>
  <cpu mode='custom' match='exact' check='full'>
    <model fallback='forbid'>Skylake-Client-IBRS</model>
    <vendor>Intel</vendor>
    <feature policy='require' name='ss'/>
    <feature policy='require' name='vmx'/>
    <feature policy='require' name='hypervisor'/>
    <feature policy='require' name='tsc_adjust'/>
    <feature policy='require' name='clflushopt'/>
    <feature policy='require' name='umip'/>
    <feature policy='require' name='md-clear'/>
    <feature policy='require' name='stibp'/>
    <feature policy='require' name='arch-capabilities'/>
    <feature policy='require' name='ssbd'/>
    <feature policy='require' name='xsaves'/>
    <feature policy='require' name='pdpe1gb'/>
    <feature policy='require' name='ibpb'/>
    <feature policy='require' name='amd-stibp'/>
    <feature policy='require' name='amd-ssbd'/>
    <feature policy='require' name='skip-l1dfl-vmentry'/>
    <feature policy='require' name='pschange-mc-no'/>
    <feature policy='disable' name='hle'/>
    <feature policy='disable' name='rtm'/>
    <feature policy='disable' name='mpx'/>
  </cpu>
  <clock offset='utc'>
    <timer name='rtc' tickpolicy='catchup'/>
    <timer name='pit' tickpolicy='delay'/>
    <timer name='hpet' present='no'/>
  </clock>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <pm>
    <suspend-to-mem enabled='no'/>
    <suspend-to-disk enabled='no'/>
  </pm>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/home/brendan/src/hubris-stuff/xenial-server-cloudimg-amd64-disk1.img' index='2'/>
      <backingStore/>
      <target dev='vda' bus='virtio'/>
      <alias name='virtio-disk0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x04' function='0x0'/>
    </disk>
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='/home/brendan/src/hubris-stuff/ubuntu.iso' index='1'/>
      <backingStore/>
      <target dev='hda' bus='ide'/>
      <readonly/>
      <alias name='ide0-0-0'/>
      <address type='drive' controller='0' bus='0' target='0' unit='0'/>
    </disk>
    <controller type='usb' index='0' model='ich9-ehci1'>
      <alias name='usb'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x7'/>
    </controller>
    <controller type='usb' index='0' model='ich9-uhci1'>
      <alias name='usb'/>
      <master startport='0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0' multifunction='on'/>
    </controller>
    <controller type='usb' index='0' model='ich9-uhci2'>
      <alias name='usb'/>
      <master startport='2'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x1'/>
    </controller>
    <controller type='usb' index='0' model='ich9-uhci3'>
      <alias name='usb'/>
      <master startport='4'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x2'/>
    </controller>
    <controller type='pci' index='0' model='pci-root'>
      <alias name='pci.0'/>
    </controller>
    <controller type='ide' index='0'>
      <alias name='ide'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x1'/>
    </controller>
    <interface type='network'>
      <mac address='52:54:00:be:c2:45'/>
      <source network='default' portid='45d91da8-2851-4e74-a63f-80ff2037e649' bridge='virbr0'/>
      <target dev='vnet0'/>
      <model type='virtio'/>
      <alias name='net0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
    </interface>
    <serial type='pty'>
      <source path='/dev/pts/3'/>
      <target type='isa-serial' port='0'>
        <model name='isa-serial'/>
      </target>
      <alias name='serial0'/>
    </serial>
    <console type='pty' tty='/dev/pts/3'>
      <source path='/dev/pts/3'/>
      <target type='serial' port='0'/>
      <alias name='serial0'/>
    </console>
    <input type='mouse' bus='ps2'>
      <alias name='input0'/>
    </input>
    <input type='keyboard' bus='ps2'>
      <alias name='input1'/>
    </input>
    <memballoon model='virtio'>
      <alias name='balloon0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x0'/>
    </memballoon>
  </devices>
  <seclabel type='dynamic' model='apparmor' relabel='yes'>
    <label>libvirt-8902ecd3-a6e9-4e59-83d5-4c05a325f67a</label>
    <imagelabel>libvirt-8902ecd3-a6e9-4e59-83d5-4c05a325f67a</imagelabel>
  </seclabel>
  <seclabel type='dynamic' model='dac' relabel='yes'>
    <label>+64055:+108</label>
    <imagelabel>+64055:+108</imagelabel>
  </seclabel>
</domain>
`

// KVMMachine represents an ephemeral libvirt-managed KVM machine.
type KVMMachine struct {
	lv     *libvirt.Libvirt
	domain libvirt.Domain
}

// Start creates and boots up a KVMMachine.
func Start(lv *libvirt.Libvirt) (*KVMMachine, error) {
	dom, err := lv.DomainCreateXML(xml, libvirt.DomainStartValidate)
	if err != nil {
		return nil, fmt.Errorf("Creating domain: %v", err)
	}
	k := &KVMMachine{
		lv:     lv,
		domain: dom,
	}
	return k, nil
}

// WriteConsole feeds the console to the provided Writer.
// The only way to cancel it is to Destroy the KVMMachine (AFAICS go-libvirt
// doesn't provide a further way to cancel either). If we had a shutdown method,
// that might work too, not sure.
func (k *KVMMachine) WriteConsole(w io.Writer) error {
	return k.lv.DomainOpenConsole(k.domain, libvirt.OptString{"serial0"}, w, 0)
}

// Destroy unceremoniously destroys the machine.
func (k *KVMMachine) Destroy() error {
	return k.lv.DomainDestroy(k.domain)
}

// NetworkAddr returns any addresses that the machine was given on the libvirt
// virtual network.
func (k *KVMMachine) NetworkAddrs() ([]string, error) {
	ifaces, err := k.lv.DomainInterfaceAddresses(
		k.domain, uint32(libvirt.DomainInterfaceAddressesSrcLease), 0)
	if err != nil {
		return nil, fmt.Errorf("DomainInterfaceAddresses: %v", err)
	}
	var ips []string
	for _, iface := range ifaces {
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}
	}
	return ips, nil
}

var logger = log.New(os.Stdout, "hubris: ", 0)

// Waits until the gues machine has taken a DHCP lease, and returns the address
// associated with the first such lease.
func awaitDHCPLease(ctx context.Context, kvm *KVMMachine) (string, error) {
	for {
		addrs, err := kvm.NetworkAddrs()
		if err != nil {
			return "", err
		}

		if len(addrs) != 0 {
			return addrs[0], nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.NewTimer(50 * time.Millisecond).C:
		}
	}

}

// Uses a hard-coded SSH config to connect to the given network address.
func dialSSH(addr string) (*ssh.Client, error) {
	// Load keys.
	path := "/home/brendan/.ssh/id_rsa"
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Reading SSH private key from %v: %v", path, err)
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Parsing SSH private key from %v: %v", path, err)
	}

	// Dial.
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("Dial SSH: %v", err)
	}
	return client, nil
}

func run(ctx context.Context) error {
	// This dials libvirt on the local machine, but you can substitute the first
	// two parameters with "tcp", "<ip address>:<port>" to connect to libvirt on
	// a remote machine.
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer func() {
		if err := l.Disconnect(); err != nil {
			logger.Printf("Disconnect: %v", err)
		}
	}()

	// We'll use this Group to do stuff that runs until kvm gets destroyed,
	// so we create it now in order to have defers lined up properly.
	eg, _ := errgroup.WithContext(context.Background())
	defer eg.Wait()

	kvm, err := Start(l)
	if err != nil {
		return err
	}
	defer func() {
		if err := kvm.Destroy(); err != nil {
			logger.Printf("Destroying KVM machine: %v", err)
		}
	}()

	eg.Go(func() error { return kvm.WriteConsole(os.Stderr) })

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	addr, err := awaitDHCPLease(ctx2, kvm)
	if err != nil {
		return fmt.Errorf("Getting guest network address: %v", err)
	}

	time.Sleep(10 * time.Second)
	client, err := dialSSH(addr + ":22")
	if err != nil {
		return err
	}
	client.Close()

	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Done")
}
