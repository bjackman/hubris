package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	"golang.org/x/sync/errgroup"
)

var xml = `
<domain type='kvm'>
  <name>instance-1</name>
  <memory unit='GiB'>1</memory>
  <os>
    <type>hvm</type>
  </os>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
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
    <controller type='pci' index='0' model='pci-root'>
      <alias name='pci.0'/>
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
  </devices>
</domain>
`

var logger = log.New(os.Stdout, "hubris: ", 0)

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

	dom, err := l.DomainCreateXML(xml, libvirt.DomainStartValidate)
	if err != nil {
		return fmt.Errorf("Creating domain: %v", err)
	}

	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		logger.Printf("Opening console")
		err := l.DomainOpenConsole(dom, libvirt.OptString{"serial0"}, os.Stderr, 0)
		logger.Printf("DomainOpenConsole returned %v", err)
		return err
	})

	<-ctx.Done()

	logger.Printf("Destroying domain..")
	if err := l.DomainDestroy(dom); err != nil {
		logger.Printf("Destroying domain %+v: %v", dom, err)
	} else {
		logger.Printf("DomainDestroy successful")
	}

	if err := l.Disconnect(); err != nil {
		logger.Printf("Disconnect: %v", err)
	}

	return eg.Wait()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		logger.Fatal(err)
	}
}
