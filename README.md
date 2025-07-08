# openav microservice biamp tesira dsp

OpenAV microservice for the Biamp Tesira Forte, Tesira Server, Tesira Serve i/o, and Tesira forteX.  Uses the microservice framework and telnet to communicate with the Biamp.

The Biamp Tesira DSP is an audio amplifier and audio switcher that utilizes Dante (Digital Audio Network Through Ethernet) protocol to transmit audio over Ethernet networks. The Biamp DSP has more than one network interface.  Its microservice sends control signals via Telnet from the orchestrator to one Biamp DSP network interface, while the audio (Dante) signals are transmitted through another network interface.  USB audio is also supported by the device.

[TesiraFORTÉ DAN CI](https://products.biamp.com/product-details/-/o/ecom-item/911.0447.900/category/FE2B76B5-8575-4F44-87A5-740FA868662F%7C1FA10A0F-C874-4DCD-B041-3833A8B78ABC%7C204E989F-7D8B-4FB6-9BDD-C5B7739EBB65)

[Microservice curl test documentation](https://github.com/Dartmouth-OpenAV/documentation/blob/main/curl_test_readme.md)

![](https://github.com/Dartmouth-OpenAV/microservice-biamp-tesira-dsp/blob/main/front.png?raw=true)
![](https://github.com/Dartmouth-OpenAV/microservice-biamp-tesira-dsp/blob/main/rear.png?raw=true)
