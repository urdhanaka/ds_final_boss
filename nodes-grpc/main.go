package main

import (
	"fmt"
	"nodes-grpc/services/virtualization"
)

func main() {
    incusCommand()
}

func libvirtCommand() {
	libvirtClient := virtualization.NewLibvirtVirt()

    err := libvirtClient.CreatePool()
    if err != nil {
        fmt.Println("Create pool error : ", err.Error())
    }

    err = libvirtClient.CreateVolume()
    if err != nil {
        fmt.Println("Create vol error : ", err.Error())
    }

    err = libvirtClient.CreateVM()
    if err != nil {
        fmt.Println("Create vm error : ", err.Error())
    }
}

func incusCommand() {
    incusClient := virtualization.NewIncusVirtualization(true)

    err := incusClient.CreateNewVM()
    if err != nil {
        fmt.Println("Create vm error : ", err.Error())
    }
}
