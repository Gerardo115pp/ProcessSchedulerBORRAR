package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/inancgumus/screen"
)

type CounterProcess struct {
	uuid           int
	counter_value  int
	process_master *ProcessList
}

func (self *CounterProcess) isAlive() (response bool) {
	self.process_master.in_use.Lock()

	defer self.process_master.in_use.Unlock()

	select {
	case someuuid := <-self.process_master.stoping_channel:
		if someuuid == self.uuid {
			response = false
			self.process_master.responses_channel <- self.uuid
		} else {
			response = true
			self.process_master.stoping_channel <- someuuid
		}
	default:
		response = true
	}
	return
}

func (self *CounterProcess) run() {
	go func(cp *CounterProcess) {
		for {

			if !cp.isAlive() {
				return
			}

			if cp.process_master.show_process {
				fmt.Printf("%d : %d\n", cp.uuid, cp.counter_value)
			}
			cp.counter_value++
			time.Sleep(time.Millisecond * 500)
		}
	}(self)
}

func (self *CounterProcess) init(uuid int, process_master *ProcessList) {
	self.uuid = uuid
	self.counter_value = 0
	self.process_master = process_master
}

type ProcessList struct {
	in_use            sync.Mutex
	stoping_channel   chan int
	responses_channel chan int
	show_process      bool
	auto_increment    int
	processes         []*CounterProcess
}

func (self *ProcessList) init() {
	self.show_process = false
	self.auto_increment = 0
	self.stoping_channel = make(chan int)
	self.responses_channel = make(chan int)
}

func (self *ProcessList) toggleDisplayFlag() {
	self.show_process = !self.show_process
}

func (self *ProcessList) scheduleProcess() {
	var new_process *CounterProcess = new(CounterProcess)
	new_process.init(self.auto_increment, self)
	new_process.run()
	self.processes = append(self.processes, new_process)

	self.auto_increment++
}

func (self *ProcessList) SendTermSingal(process_uuid int) {
	self.stoping_channel <- process_uuid
	var reciver_uuid int = <-self.responses_channel

	if reciver_uuid == process_uuid {
		fmt.Println("Process Stopped!")
	} else {
		fmt.Printf("Got response from a diferent process with uuid = %d", reciver_uuid)
	}
}

func clear() {
	screen.Clear()
	screen.MoveTopLeft()
}

func runMenu() {
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	var process_list *ProcessList = new(ProcessList)
	var choice string
	var process_uuid_holder int
	process_list.init()
	for {
		if !process_list.show_process {
			clear()
		}
		fmt.Println("1 - Agregar Proceso\n2 - Mostrar Proceso\n3 - Eliminar Proceso\n4 - Salir")
		scanner.Scan()
		choice = scanner.Text()
		switch choice {
		case "1":
			process_list.scheduleProcess()
		case "2":
			process_list.toggleDisplayFlag()
		case "3":
			fmt.Printf("uuid del proceso>>> ")
			scanner.Scan()
			process_uuid_holder, _ = strconv.Atoi(scanner.Text())
			process_list.SendTermSingal(process_uuid_holder)
		case "4":
			return
		}
	}
}

func main() {
	runMenu()
}
