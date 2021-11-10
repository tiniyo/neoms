<div id="top"></div>
<!--
*** Thanks for checking out the Neoms - best opensource alternative to twilio platform. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/tiniyo/neoms">
    <img src="images/logo.png" alt="Logo" width="300" height="80">
  </a>

<h3 align="center">NeoMs</h3>

  <p align="center">
     The open-source alternative to Twilio.
    <br />
    <a href="https://tiniyo.com/dist/index-v1.html?version=v1"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://tiniyo.com">View Demo</a>
    ·
    <a href="https://github.com/tiniyo/neoms/issues">Report Bug</a>
    ·
    <a href="https://github.com/tiniyo/neoms/issues">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

[![Product Name Screen Shot][product-screenshot]](https://tiniyo.com)

Project NeoMs is open-source alternative to twilio voice api. It helps software developer or enterprises to build CPaaS like twilio with their infrastructure.

<p align="right">(<a href="#top">back to top</a>)</p>



### Built With

* [FreeSWITCH](https://github.com/signalwire/freeswitch)
* [GoLang](https://golang.org/)
* [Redis](https://redis.io/)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

NeoMs required local redis on each server where you setup NeoMs. Redis is used for live call stats.

* Redis
  ```sh
  sudo apt install redis-server
  ```

* FreeSWITCH (You may install your desired version)
  ```sh
  sudo apt-get update && apt-get install -y gnupg2 wget lsb-release
  wget -O - https://files.freeswitch.org/repo/deb/debian-release/fsstretch-archive-keyring.asc | apt-key add -
  echo "deb http://files.freeswitch.org/repo/deb/debian-release/ `lsb_release -sc` main" > /etc/apt/sources.list.d/freeswitch.list
  echo "deb-src http://files.freeswitch.org/repo/deb/debian-release/ `lsb_release -sc` main" >> /etc/apt/sources.list.d/freeswitch.list
  # you may want to populate /etc/freeswitch at this point.
  # if /etc/freeswitch does not exist, the standard vanilla configuration is deployed
  # apt-get update && apt-get install -y freeswitch-meta-all
  ```
  
### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/tiniyo/neoms.git
   ```
2. Install the neoms package dependency using go mod. 
   ```sh
   cd neoms
   go mod download
   ```
3. Please enter your api configurations using environment variable.
   ```sh
   EXPORT REGION=ALL
   EXPORT SER_USER=API_BASIC_AUTH_USER
   EXPORT SER_SECRET=API_BASIC_AUTH_SECRET
   EXPORT SIP_SERVICE=SIP_SERVICE_URL
   EXPORT KAMGO_SERVICE=LOCATION_SERVICE_URL
   EXPORT NUMBER_SERVICE=NUMBER_SERVICE_URL
   EXPORT HEARTBEAT_SERVICE=HEARTBEAT_SERVICE_URL
   EXPORT RATING_ROUTING_SERVICE=RATING_ROUTING_SERVICE_URL
   EXPORT CDR_SERVICE=CDR_SERVICE_URL
   EXPORT RECORDING_SERVICE=RECORDING_SERVICE_URL
   ```
4. build the neoms and run the n
```sh
go build -o neoms main.go
./neoms
```

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

_For more examples, please refer to the [Documentation](https://tiniyo.com)_

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [X] XML Parser
- [X] Play, Say, Dial, Sip, Number, Record, Hangup, Reject, Pause, Gather , Redirect Twilio elements support
- [] Conference
    - [] Conference API
    - [] Inbound XML Support
    - [] Outbound XML Support
- [X] HeartBeat Event for billing
- [X] CDR Post to api
- [X] Initiated, Ringing, In-Progress, Hangup, Record Start, Record Stop Event for Twilio based callback for customer Url
- [X] HeartBeat Event for billing
- [X] Text2Speech Support
- [] Speech2Text Support
- [X] DTMF Support
- []  Speech Detection

See the [open issues](https://github.com/tiniyo/neoms/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Your Name - [@twitter_handle](https://twitter.com/twitter_handle) - support@tiniyo.com

Project Link: [https://github.com/tiniyo/neoms](https://github.com/tiniyo/neoms)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* []()
* []()
* []()

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/tiniyo/neoms.svg?style=for-the-badge
[contributors-url]: https://github.com/tiniyo/neoms/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/tiniyo/neoms.svg?style=for-the-badge
[forks-url]: https://github.com/tiniyo/neoms/network/members
[stars-shield]: https://img.shields.io/github/stars/tiniyo/neoms.svg?style=for-the-badge
[stars-url]: https://github.com/tiniyo/neoms/stargazers
[issues-shield]: https://img.shields.io/github/issues/tiniyo/neoms.svg?style=for-the-badge
[issues-url]: https://github.com/tiniyo/neoms/issues
[license-shield]: https://img.shields.io/github/license/tiniyo/neoms.svg?style=for-the-badge
[license-url]: https://github.com/tiniyo/neoms/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://in.linkedin.com/company/tiniyo
[product-screenshot]: images/screenshot.png
