package main

import (
  "fmt"
  "time"
  "strconv"
  "github.com/stianeikeland/go-rpio/v4"
)

var input string
var exit bool = false
var startStop bool = false
var speed time.Duration = 1  //subtract time by speed to speed up servo movements
var count int

//Positions will be indexed [left/right][up/down]
//index 0 will be starting position
var servoPositions [][]uint32 = [][]uint32{
  []uint32{2400, 2400}, //all the way left and up
  []uint32{2200, 2100}, //thumb
  []uint32{1800, 2100}, //first finger
  []uint32{1600, 2100}, //second
  []uint32{1500, 2100}, //third
  []uint32{1350, 2100}, //outside pinky
}

func checkUserInput(){
  for{
    // Take standard input from user (maybe use gpio button input eventually)
    fmt.Scanln(&input)

    if input == "exit" {
      exit = true
      break
    }else if input == "start" {
      startStop = true
    }else if input == "stop" {
      startStop = false
    }else if len(input) > 6 {
      if input[:6] == "speed=" {
        num, err := strconv.Atoi(input[6:]) //everything after "="
        if err == nil {
          speed = time.Duration(num) * time.Millisecond
          fmt.Println("Speed set to: ", speed)
        }else{
          fmt.Println("Invalid Speed")
        }
      }
    }
  }
}

func main() {
  fmt.Println("Servo Time!")
  fmt.Println("Commands: ")
  fmt.Println("-\"exit\" exits program and cleans up GPIO.")
  fmt.Println("-\"speed=x\" sets speed x for game to run at.")
  fmt.Println("-\"start\" starts or resumes servos and game.")
  fmt.Println("-\"stop\" pauses program.")

  go checkUserInput()


  //Open memory range for GPIO access in /dev/mem
  err := rpio.Open()
  if err != nil {
    panic(err) //panic logs error and exits program
  }
  // Unmap gpio memory when main() exits
  defer rpio.Close()
  //After all done, stop PWMing
  defer rpio.StopPwm()

  //PWM for servo control
  //Pin Numbers are GPIO numbers
  servoLR := rpio.Pin(19)  //servo Left Right rotation physical pin 35
  servoUD := rpio.Pin(18)  //servo Up Down rotation physical pin 12

  servoLR.Mode(rpio.Pwm)
  servoUD.Mode(rpio.Pwm)

  servoLR.Freq(1200000) //1.2MHz
  servoUD.Freq(1200000)

  //1200000 / 24000 = 50Hz servo PWM period speed
  //From SG90 datasheet: 1ms pulse width = -90 degrees, 2ms pulse = 90 degrees
  //1ms to 2ms with 20ms(50Hz) period are pretty standard servo timings
  //1ms / 20ms = 0.05, 0.05 * 24000 = 1200 = -90 degrees
  //2ms / 20ms = 0.1, 0.1 * 24000 = 2400 = 90 degrees

  //Set Duty Cycle
  //params: dutyLength, cycleLength
  servoLR.DutyCycle(1800, 24000)
  servoUD.DutyCycle(1800, 24000)

  for !exit {
    if startStop {
      //Do the servo knife game
      time.Sleep(time.Millisecond * 250 - speed)

      servoLR.DutyCycle(servoPositions[count][0], 24000)
      time.Sleep(time.Millisecond * 10)
      servoUD.DutyCycle(servoPositions[count][1], 24000)

      time.Sleep(time.Millisecond * 250 - speed)
      //go back to start position
      servoUD.DutyCycle(servoPositions[0][1], 24000)
      time.Sleep(time.Millisecond * 10)
      servoLR.DutyCycle(servoPositions[0][0], 24000)

      count++

      //make sure count doesn't overflow
      if count >= len(servoPositions) {
        count = 0
      }
    }else{
      //stopped
      //(sleep to give processor time to run other goroutines and save power)
      time.Sleep(time.Millisecond)
    }
  }

  //After exit == true
  rpio.StopPwm()
}
