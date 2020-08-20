# Modbus "Gateway" Plugin

The Modbus Gateway plugin collects Input Registers and Holding
Registers via Modbus TCP.  It is similar to the "modbus" driver (and uses the same
underlying protocol implementation) but has a different configuration format suited to
communicating with Modbus/TCP devices acting as _gateways_. A gateway is still a modbus server,
but instead of reaching one device a gateway has many real or virtual devices attached to it.
The devices beyond the gateway are often connected via modems, radios, or multi-drop RS-485
or RS-422 buses.

The motivation behind this plugin is to address the situation where there are multiple devices
behind a gateway, but the gateway can't accept many TCP connections at once.  This is
a typical case (unfortunately) because many gateways on the market are made using
small microcontrollers with small RAM and limited TCP/IP stacks.  Even some
expensive devices, like the PowerLogic CM4200, only allow 4 simultaneous connections.
The [MBAP](https://modbus.org/docs/Modbus_Messaging_Implementation_Guide_V1_0b.pdf)
protocol includes a _unit identifier_ in each request specifically so that one TCP
connection can be shared talking to multiple devices.  (Historical note: Modbus/UDP exists
as well, but is used less often).

### Naming Conventions

To address the Modbus Organization's
[new naming conventions](https://modbus.org/docs/Client-ServerPR-07-2020-final.docx.pdf)
we are adopting the following terminology:

 - _Address_ refers to a 16-bit register address.  Note that _input_ and _holding_ registers
    occupy different address spaces (you can have an input register 1000 and a holding
    register 1000, and they are different)
 - _Gateway_ is a modbus "Server" thay _may_ have multiple devices attached to it
 - _Unit Address_ is the 8-bit address of an individual physical or virtual device
    attached to the gateway.  It is difficult to think of a microcontroller with a
    serial port and 2KB of RAM as a "server" so in agreement with the
    [MBAP](https://modbus.org/docs/Modbus_Messaging_Implementation_Guide_V1_0b.pdf)
    specification I'm calling it a "device" that has a _Unit Address_.


## When to use this plugin

 - The Modbus/TCP device answers to multiple unit addresses and you
   do not want the plugin to make separate, concurrent TCP connections to each  
 - There are multiple modbus devices attached to the gateway (typically serial),
   each with it's own distinct unit address
 - You want more control over how register fetches are grouped into
   bulk modbus requests (even when the registers you want to fetch are not
   immediately adjacent to each other)

## When to use the original modbus plugin

 - The device uses a direct serial connection (RS-232, RS-485)
 - The request is for modbus functions other than _READ INPUT_ and
   _READ MULTIPLE HOLDING REGISTERS_.  This plugin does not support discretes
   (digital inputs and outputs) at this time (planned to be added later)
 - The data types being retrieved are other than INT16, UINT16, INT32, UINT32
 - The device uses ASCII mode (rare these days)
  
 ## Configuration

Each `[[input]]` communicates to one gateway and one or more devices.  It is
perfectly reasonable to have only one device on the gateway - in fact, this
is the typical case when you have a single Ethernet modbus device.

In the configuration you create _requests_ then map the results of those requests
to _fields_.  A request defines a single modbus _PDU_ (protocol data unit), or one message
sent to the gateway that (hopefully) solicits exactly one response message.  The payload of
that response is a string of bytes.  _Field_ definitions tell the plugin how to interpret
those bytes to turn the received values into measurements.

Sometimes a modbus device has a register layout like this:

`[1000][1001][1002][1003]`

but you don't want to store the middle registers at all.  Why would you do that? Because it
can be less "expensive" to have fewer requests.  If you were to request all four values,
but only assign 1000 and 1993 to measurements, you sent 8 unwanted bytes across the
wire.  That's far cheaper than requesting 1000 and 1003 in totally separate requests.
To accomodate this type of request a field can be marked _omit=true_ meaning the
response is received but not stored as a field of the measurement. 

## Sample Configuration
```toml
[[inputs.modbusgw]]
    #
    # Name of this input - should be unique
    #
    name="sma"

    #
    # Address and port of the modbus server or gateway
    #
    gateway="tcp://yourserver.com:502"

    #
    # Response timeout, in go duration fornat
    # Usually can be set pretty low
    #
    timeout="5s"

    #
    # Request (poll) definitions
    #
    # Request parameters:
    #
    # unit - required.  Unit address of device being polled.  Per spec, the value is between
    #    1 and 247, or 0 for broadcast.  The values 0 or 255 are usually accepted to communicate
    #    directly to a Modbus/TCP device not acting as a gateway.  Historically this has been a
    #    point of confusion.  If this is what you want (to talk to the gateway itself), try 255 first.
    #    Using the broadcast address can cause unexpected device responses.
    #
    # address - the register address of the first register being requested.  This address is zero-based.
    #    For example, the first holding register is address 0.  Be aware that some documentation
    #
    # count - how mant 16-bit registers to request
    #
    # type - defines the register type, which maps internally to the function code used ub the
    #   PDU (request).  Must be "holding" or "input", if unspecified defaults to "holding"
    #
    # measurement - the nameof the measurement, for example when stored in influx
    #
    # fields - defines how the response PDU is mapped to fields of the measurement.  Attributes
    # of each field are:
    #
    # name - name of the field
    #
    # type - must be INT32, UINT32, INT16, or UINT16.  More tyoes will be added in the future.
    #
    # scale, offset - math performed on the raw modbus value before storing.
    #    stored field value = (modbus value * scale) + offset
    #
    # omit - if true, don't store this field at all.  you must still set a 'type'.  Use this to
    #   skip fields not of interest that are part of the response because they are within the
    #   requested register range.
    #
    requests = [
        { unit=3, address=30769, count=8, type="holding", measurement="pv1", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=4, address=30769, count=8, type="holding", measurement="pv2", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=5, address=30769, count=8, type="holding", measurement="pv3", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=6, address=30769, count=8, type="holding", measurement="pv4", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=7, address=30769, count=8, type="holding", measurement="pv5", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=8, address=30769, count=8, type="holding", measurement="pv6", fields = [
            {name="Ipv", type="INT16", scale=0.001},
            {name="Vpv", type="INT16", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=9, address=30769, count=8, type="holding", measurement="pv7", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
     ]
```