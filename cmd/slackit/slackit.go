///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	type Payload struct {
		IconEmoji   string `json:"IconEmoji,omitempty"`
		IconURL     string `json:"IconURL,omitempty"`
		Text        string `json:"Text,omitempty"`
		ChannelName string `json:"ChannelName,omitempty"`
		Color       string `json:"Color,omitempty"`
	}

	icebreakers := [...]string{
		"When someone finds out what you do, or where you are from, what question do they always ask you?",
		"What’s the best thing you’ve got going on in your life at the moment?",
		"What scene in a movie always gives you goosebumps every time you watch it?",
		"What incredibly common thing have you never done?",
		"What topic could you give a 20-minute presentation on without any preparation?",
		"What are some of your favorite games to play?",
		"What takes a lot of time but is totally worth it?",
		"What is the most amazing fact you know?",
		"What website or app doesn’t exist, but you really wish it did?",
		"What’s your favorite type of day? (weather, temp, etc.)",
		"What is the most clever or funniest use of advertising you’ve seen?",
		"How into self-improvement are you?",
		"Are you more productive at night or in the morning? Do you think it’s possible to change and get used to another schedule?",
		"What has taken you the longest to get good or decent at?",
		"What food do you love that a lot of people might find a little odd?",
		"If you could start a charity, what would it be for?",
		"What was the funniest thing you’ve seen recently online?",
		"What’s something that a lot of people are missing out on because they don’t know about it?",
		"What are some of your guilty pleasures?",
		"Who is the most interesting person you’ve met and talked with?",
		"What has really taken a toll on you?",
		"How did you spend the money from your very first job?",
		"What do you wish someone taught you a long time ago?",
		"Do you think you rely too heavily on your phone? Why or why not?",
		"How could carousels be spiced up so they are more exciting?",
		"What’s your favorite car that you’ve owned?",
		"What subjects should be taught in school but aren’t?",
		"What’s the biggest vehicle you’ve driven?",
		"What songs would be played on a loop in hell?",
		"What rule do you wish they would introduce into your favorite sport?",
		"What kind of challenges are you facing these days?",
		"What do you highly recommend to most people you meet?",
		"Do you think you have a pretty good work-life balance? Why or why not?",
		"What was the last thing you were really excited about?",
		"What does your perfect breakfast look like?",
		"What are some of your favorite holiday traditions that you did while growing up?",
		"If you could choose your dreams, what would you prefer to dream about?",
		"Would you ride in a zeppelin if given a chance?",
		"What’s something that was once important but is now becoming less and less relevant?",
		"What tells you the most about a person?",
		"When is the most interesting period in history?",
		"What is the best pair of shoes you have owned?",
		"What book had the most significant impact on you?",
		"Where’s your favorite place to nap?",
		"What is the best event you’ve attended?",
		"What do you buy way more of than most people?",
		"What do you rebel against?",
		"What well-known person does the most good for the world?",
		"What’s your favorite food combination?",
		"What’s the weirdest way you have met someone?",
		"What’s the most amazing natural occurrence you’ve witnessed?",
		"How did you get that scar of yours?",
		"What do you wish was illegal?",
		"If someone came up to you and said “Hey, do that thing you do!”, what thing would pop into your head first?",
		"Who is the most intelligent or creative person you know?",
		"What wastes the most time in your day to day life?",
		"What’s a problem you have, that might be entirely unique to you?",
		"What company or brand did you love until they betrayed your trust?",
		"Would you ever try space tourism, if you had the money for it?",
		"What’s the most annoying machine you must deal with regularly?",
		"What are you grateful for?",
		"What hobby would you be a lot of fun to get into?",
		"What do you resent paying for most?",
		"What pets did you have growing up?",
		"What’s the best or worst prank you’ve played on someone?",
		"What was the scariest movie you’ve seen?",
		"What motivates you?",
		"Where are five places you really want to visit before you die?",
		"What’s the best location to fully enjoy a good cup of coffee?",
		"How handy are you when it comes to fixing things?",
		"What skill or talent would you most like to learn?",
		"What weird thing do you have nostalgia for?",
		"What works of art have really made an impression on you?",
		"What culture would you like to learn more about?",
		"If you were featured on the local news, what would you most likely be on there for?",
		"What do you wish your phone could do?",
		"What’s your favorite international food?",
		"What workers have the worst jobs?",
		"What kind of physical activities do you like doing?",
		"Would you rather watch a movie on your TV at home or on the big screen in the theater, and why?",
		"What assumption you made went hilariously wrong?",
		"What big problem do you think technology will solve next?",
		"What fashion trend needs to be brought back?",
		"How do you think you will be/act when you are old?",
		"What’s your favorite way to waste time online?",
		"What’s the longest trip you’ve been on?",
		"What was something you thought would be easy until you tried it?",
		"What, in your opinion, is the most amazing animal?",
		"How into tech are you? Why?",
		"What is the biggest mistake you’ve made at work?",
		"Who is the oldest person you know personally? What interesting stories have they told you?",
		"Who is the funniest person in your family?",
		"What useless facts do you know?",
		"What’s your favorite Olympic sport to watch?",
		"How do you usually get your news?",
		"What smell do you hate that doesn’t seem to bother other people?",
		"What’s the most delightful hotel or house you’ve stayed in on vacation?",
		"What weird quirks did you pick up from your parents?",
		"What were your favorite television shows when you were growing up?",
		"What’s the silliest thing you are pretty good at?",
		"What’s your idea of a great day?",
		"What are some of the dumbest misadventures you’ve been on?",
		"What do you want to do when you retire?",
		"What do you do to unwind after a hard day?",
		"Who is the most competitive person you know?",
		"Would you rather spend time with other people or at home alone?",
		"What are some films that would make it on to your top 50 list of movies?",
		"What slang are you really happy went out of fashion?",
		"What music do you put on when you want to get pumped?",
		"What makes you feel old when you think about it?",
		"How well do you know your neighbors?",
		"What’s the best backhanded compliment you’ve heard / can think of on the spot?",
		"Who is the most interesting stranger you’ve met?",
		"What was your funniest or worst experience with a dentist?",
		"What’s the noblest endeavor a person can dedicate their life to?",
		"What’s your idea of a great party?",
		"What are some of your favorite scenes from movies?",
		"What are two of your favorite snacks?",
		"What’s the biggest adventure you’ve been on?",
		"Besides war and diplomacy, what would be the best way for countries to settle disputes?",
		"What household chore do you actually enjoy?",
		"What is something you do better than most people and something you do worse than most people?",
		"What TV show are you hooked on or were recently hooked on?",
		"If you had to lose one of your senses, which would you choose to lose?",
		"If a new volcano formed and the government had an online contest to see what it would be named, what name would you submit?",
		"What is the last goal you achieved?",
		"What was your worst haircut experience?",
		"Do you prefer pens or pencils? Why?",
		"What’s the best way to stay young?",
		"What is the most stressful TV show or movie you watched?",
		"How good are you at drawing?",
		"What did your teachers and parents say would be really important when you grew up, but it hasn’t been? ",
		"How do you feel about clowns?",
		"What is your favorite genre of movie? Why?",
		"What brands do you love/hate the most?",
		"What is the weirdest food combination you’ve made and tried?",
		"What would you change if you were in charge of the company you work for?",
		"Who has been your most interesting/confusing/annoying neighbor?",
		"Has there ever been a time when something so amazing or unexpected happened that it literally left you speechless for a time?",
		"Where’s the most surreal area you been to?",
		"What are common misconceptions about your job?",
		"What public spaces do you feel most comfortable in? (Library, bar, park, mall, stadium, etc.)",
		"What’s the most outdated piece of tech you still use regularly?",
		"What’s the weirdest food you’ve eaten?",
		"Who was your favorite teacher?",
		"What recent trend are you totally on board with?",
		"What about becoming an adult caught you completely off guard?",
		"How often do you dance?",
		"What’s the best cake you’ve ever eaten?",
		"What kind of art do you appreciate the most?",
		"What crossed way too far into the uncanny valley for you?",
		"What’s your favorite thing about the area/city/state you live in?",
		"What’s the best day you’ve had recently?",
		"What’s your favorite way to spend time outdoors?",
		"What does your perfect burger or sandwich have in it?",
		"What’s the worst advice you’ve been given?",
		"What’s the strangest name someone you have met had?",
		"What was something courageous you’ve (in person) seen someone do?",
		"What card or board games do you like to play?",
		"What are you best at fixing?",
		"What movie never gets old no matter how many times you’ve seen it?",
		"If the universe is just a simulation, what update or patch does it need?",
		"Where have you traveled to?",
		"What’s the scariest horror movie or horror book monster?",
		"What’s the most unique shop or restaurant you’ve been in?",
		"What hard time in your life left you a better person after it was finished?",
		"What’s the best sports game you’ve been to?",
		"What period would be the best to be born in?",
		"What sport do you wish you knew more about?",
		"What’s something you’re looking forward to?",
		"What’s the most creative thing you’ve done?",
		"What are you hilariously bad at?",
		"Tell me about a time you were totally out of your element/comfort zone.",
		"Are you a cat or dog person or neither? Why?",
		"Who is the most gifted person you know?",
		"What was your best vacation?",
		"What do you usually do on your commute to work?",
		"What was the craziest theme park or fair ride you’ve been on?",
		"What kind of people do you most enjoy hanging out with?",
		"What do you think the ideal age to be is?",
		"What are you kind of snobby about?",
		"What toy did you hate most as a child?",
		"What’s your drink of choice? (Either alcoholic or non.)",
		"If you left your current life behind and ran away to follow your dreams, what would you be doing?",
		"What food is underrated or underappreciated?",
		"What’s your favorite line from a book or a movie?",
		"What is the best thing you have ever bought?",
		"What catchy jingle or bit of advertising has stuck with you all these years?",
		"What luxury is totally worth the price?",
		"What is the most unique or silliest problem you have going on in your life at the moment?",
		"What could movie theaters do to improve the experience of going there?",
		"If you were so wealthy you didn’t need to work, what would you do with your time?",
		"What apps do you use most?",
		"What is the most tedious and/or the most exciting sport to watch?",
		"What’s your favorite island that you’ve visited?",
		"What do you geek out about?",
		"Besides insects and spiders, what animals annoy you the most?",
		"What’s your favorite month?",
		"What’s the best concert you’ve been to and why was it so good?",
	}
	rand.Seed(time.Now().Unix())
fmt.Println(icebreakers[rand.Intn(len(icebreakers))])
	data := Payload{
		IconEmoji:   ":interrobang:",
		IconURL:     "https://slack.global.ssl.fastly.net/9fa2/img/services/hubot_128.png",
		Text:        "sun: sup",
		ChannelName: "1s-and-0s-deploys", // "khan-district-eng",
		Color:       "#3AA3E3",
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "http://localhost:8080/slack", body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bytey, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(bytey))
}
