package main

import "fmt"
import "math"
import "encoding/binary"
import "os"



var Shift = 0.0
var Strech = uint(1)
var source = 0
var sound []float64
var rate = 1.0
var offset = 0
var multiplier = 1.0

func load(s string){
	sound = make([]float64, 48000*10)
	fin, _ := os.Open(s)
	for i:=0; i<len(sound); i++{
		_, err := fmt.Fscanln(fin, &(sound[i]))
		if err != nil {
			sound = sound[0:i]
			return
		}
	}
}

func get(i float64)float64{
	if source == 0 {
		return math.Sin(i*2*math.Pi)
	}
	pos := i*48000/rate
	x := sound[int(math.Floor(pos))+offset]
	y := sound[int(math.Floor(pos))+offset+1]
	d := pos - math.Floor(pos)
	return multiplier*(x+d*(y-x))
}


//p pitch 0 = A4 = 440Hz; 12 = A5 ~ 880Hz
//l length 44100 = 1s
//s style 0=empty, 1=normal, 2=trill, 3=start, 4=stop, 5=startstop, others=normal
func note(p1 float64, p2 float64, l uint, s int) ([]float64,[]float64,[]float64){
	rst1 := make([]float64, l)
	rst2 := make([]float64, l)
	rst3 := make([]float64, l)
	p1 = p1+Shift
	p2 = p2+Shift
	for i:=uint(0); i<l; i++{
		rst2[i] = 0.0
		rst3[i] = 1.0
		if s==0 {
			rst1[i] = 0
			rst2[i] = 1.0
			if i < 2400 {
				rst3[i] = float64(i)/2400.0
			} else {
				rst3[i] = 0.0
			}
		} else if s==1 {
			pi := p1 + (p2-p1)*float64(i)/float64(l)
			rst1[i] = 440*math.Exp2(pi/12.0)
			rst2[i] = 1.0
		} else if s==2 {
			pi := p1 + (p2-p1)*float64(i)/float64(l) + math.Sin(2*math.Pi*8*float64(i)/48000)*0.5
			rst1[i] = 440*math.Exp2(pi/12.0)
			rst2[i] = 1.0
		} else if s==3 {
			pi := p1 + (p2-p1)*float64(i)/float64(l)
			rst1[i] = 440*math.Exp2(pi/12.0)
			rst2[i] = 0.0
			if i < 2400 {
				rst3[i] = float64(i)/2400.0
			} else {
				rst3[i] = 1.0
			}
		} else if s==4 {
			pi := p1 + (p2-p1)*float64(i)/float64(l)
			rst1[i] = 440*math.Exp2(pi/12.0)
			rst2[i] = 0.0
			if (l-i) < 2400 {
				rst3[i] = float64(l-i)/2400.0
			} else {
				rst3[i] = 1.0
			}
		} else if s==5 {
			pi := p1 + (p2-p1)*float64(i)/float64(l)
			rst1[i] = 440*math.Exp2(pi/12.0)
			rst2[i] = 0.0
			if i < 2400{
				rst3[i] = float64(i)/2400.0
			}else if (l-i) < 2400 {
				rst3[i] = float64(l-i)/2400.0
			} else {
				rst3[i] = 1.0
			}
		}
		
	}
	return rst1, rst2, rst3
}

func wave(n []float64, sty []float64, force []float64, fun func(float64)(float64)) []float64 {
	rst := make([]float64, len(n))
	cur := 0.0
	for i:=0; i<len(rst); i++{
		cur = cur + n[i]
		//if cur > 48000.0 {
		//	cur = cur - 48000.0
		//}
		if  force[i] == 0 {
			cur = 0
		}
		v := cur/48000.0//+sty[i]
		if i>1 && n[i-1]==0.0{
			cur = 0.0
			v = 0.0
		}
		rst[i] = fun(v)*force[i]*0.5
	}
	return rst
}

func parse(in string) ([]float64, []float64, []float64){
	p1 := 0.0
	p2 := 0.0
	l := uint(0)
	s := 0
	bs := []byte(in)
	idx := 0

	if bs[0] == '^' {//start
		s = 3
		idx = 1
	} else if bs[0] == '~' {//trill
		s = 2
		idx = 1
	} else if bs[0] == 'x' {//empty
		s = 0
		idx = 1
	} else if bs[0] == '|' {//stop
		s = 4
		idx = 1
	} else if bs[0] == '/' {//start-stop
		s = 5
		idx = 1
	} else if bs[0] == '#' {//comment
		return nil, nil, nil
	} else if bs[0] == '@' {//pitch shift
		tmp := string(bs[1:])
		fmt.Sscan(tmp, &Shift)
		return nil, nil, nil
	} else if bs[0] == '*' {
		tmp := string(bs[1:])
		fmt.Sscan(tmp, &Strech)
		return nil, nil, nil
	} else if bs[0] == '!' {
		tmp1 := ""
		tmp2 := ""
		tmp3 := ""
		if string(bs[1:])=="sin" {
			source = 0
			return nil, nil, nil
		} else {
			sta := 0
			for i:=1; i<len(bs); i++{
				if bs[i] == '%'{
					sta++
					continue
				}
				if sta == 0 {
					tmp1 += string(bs[i])
				} else if sta == 1 {
					tmp2 += string(bs[i])
				} else if sta == 2 {
					tmp3 += string(bs[i])
				}
			} 
		}
		if tmp1=="" || tmp2== "" || tmp3 == ""{
			return nil, nil, nil
		}
		load(tmp1)
		_, err := fmt.Sscan(tmp2, &rate)
		if err != nil {
			fmt.Fprintln(os.Stderr, string(bs), err)
			os.Exit(0)
		}
		_, err = fmt.Sscan(tmp3, &offset)
		if err != nil {
			fmt.Fprintln(os.Stderr, string(bs), err)
			os.Exit(0)
		}
		source = 1
		return nil, nil, nil
	} else if bs[0] == 'm'{
		tmp := string(bs[1:])
		fmt.Sscan(tmp, &multiplier)
		return nil, nil, nil
	} else {
		s = 1
		idx = 0
	}
	str := make([]string, 3)
	sid := 0
	for i:=idx; i<len(bs); i++{
		switch bs[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '-' :
			str[sid]+=string(bs[i])
		case '=':
			sid = 1
		case ':':
			sid = 2
		default:
			fmt.Fprintln(os.Stderr, "Error!", in)
			os.Exit(0)
		}
	}
	fmt.Sscan(str[0], &p1)
	if str[1]==""{
		p2 = p1
	} else {
		fmt.Sscan(str[1],&p2)
	}
	rl := 0.0
	fmt.Sscan(str[2],&rl)
	if rl == 1.0 {
		l = 96000
	} else if rl == 2.0 {
		l = 48000
	} else if rl == 3.0 {
		l = 32000
	} else if rl == 4.0 {
		l = 24000
	} else if rl == 6.0 {
		l = 16000
	} else if rl == 8.0 {
		l = 12000
	}else if rl == 12.0 {
		l = 8000
	}else if rl == 16.0 {
		l = 6000
	} else if rl == 24.0 {
		l = 4000
	}else if rl == 32.0 {
		l = 3000
	}else if rl == 48.0 {
		l = 2000
	}else if rl == 64.0 {
		l = 1500
	}else if rl == 96.0 {
		l = 1000
	}else{
		l = uint(rl/100.0)
	}
	return note(p1, p2, l*Strech, s)
}

func all(name string)([]float64, []float64, []float64){
	rst1 := make([]float64, 600*48000)
	rst2 := make([]float64, 600*48000)
	rst3 := make([]float64, 600*48000)
	size := 0
	f, e := os.Open(name)
	if e!=nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(0)
	}
	str := ""
	for {
		_, e = fmt.Fscan(f, &str)
		if e==nil {
			a, b, c := parse(str)
			//fmt.Println(str, len(a), size, float64(size)/48000.0)
			for i:=0; i<len(a); i++{
				rst1[size+i] = a[i]
				rst2[size+i] = b[i]
				rst3[size+i] = c[i]
			}
			size += len(a)
		} else {
			break
		}
	}
	return rst1[0:size], rst2[0:size], rst3[0:size]
}

func main(){
	if len(os.Args)<2 {
		fmt.Fprintln(os.Stderr, "arg!!")
		return
	}
	a, b, c := all(os.Args[1])
	size := len(a)*2
	x := make([]uint8, size+44)
	x[0] = 'R'
	x[1] = 'I'
	x[2] = 'F'
	x[3] = 'F'
	x[4] = uint8((size+44-8)&0xff)
	x[5] = uint8(((size+44-8)>>8)&0xff)
	x[6] = uint8(((size+44-8)>>16)&0xff)
	x[7] = uint8(((size+44-8)>>24)&0xff)
	x[8] = 'W'
	x[9] = 'A'
	x[10] = 'V'
	x[11] = 'E'
	x[12] = 'f'
	x[13] = 'm'
	x[14] = 't'
	x[15] = ' '
	x[16] = 16&0xff
	x[17] = (16>>8)&0xff
	x[18] = (16>>16)&0xff
	x[19] = (16>>24)&0xff
	x[20] = 1
	x[21] = 0
	x[22] = 1
	x[23] = 0
	x[24] = 48000&0xff
	x[25] = (48000>>8)&0xff
	x[26] = (48000>>16)&0xff
	x[27] = (48000>>24)&0xff
	x[28] = 96000&0xff
	x[29] = (96000>>8)&0xff
	x[30] = (96000>>16)&0xff
	x[31] = (96000>>24)&0xff
	x[32] = 2
	x[33] = 0
	x[34] = 16
	x[35] = 0
	x[36] = 'd'
	x[37] = 'a'
	x[38] = 't'
	x[39] = 'a'
	x[40] = uint8(size&0xff)
	x[41] = uint8((size>>8)&0xff)
	x[42] = uint8((size>>16)&0xff)
	x[43] = uint8((size>>24)&0xff)
	d := wave(a, b, c, get)
	for i:=0; i<len(d); i++{
		v16 := int16(30000.0*d[i])
		x[i*2+44] = uint8(v16&0xff)
		x[i*2+45] = uint8((v16>>8)&0xff)
	}
	fout, _ := os.Create(os.Args[1]+".wav")
	binary.Write(fout, binary.LittleEndian, x)
	fout.Close()
}
