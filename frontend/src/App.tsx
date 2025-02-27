import { use, useEffect, useState } from 'react';
import { Greet, GetAllConnectedIPs, GetPortAverage, GetLogs, GetPortRate, GetPortAverageB } from "../wailsjs/go/main/App";
import "./globals.scss";
import "./inf.scss"
import {
    Header,
    HeaderContainer,
    HeaderName,
    HeaderNavigation,
    HeaderMenuButton,
    HeaderMenuItem,
    HeaderGlobalBar,
    HeaderGlobalAction,
    SkipToContent,
    SideNav,
    SideNavItems,
    SideNavMenu,
    SideNavMenuItem,
    SideNavLink,
    SideNavDivider,
    HeaderSideNavItems,
    Content,
    Grid,

    Column,
    GlobalTheme,
    Theme

} from '@carbon/react';
import { Switcher, Notification, UserAvatar, Fade, ConditionPoint } from '@carbon/icons-react';
import { server } from '../wailsjs/go/models';


const App = () => {
    const [connectedIPs, setConnectedIPs] = useState<Record<string, server.IPConnection>>({});
    const [selectedIP, setSelectedIP] = useState<string | null>(null);
    const [selectedDevNum, setSelectedDevNum] = useState<number | null>(null);
    const [vdsAverage, setVdsAverage] = useState<number | null>(null);
    const [vgsAverage, setVgsAverage] = useState<number | null>(null);
    const [intTempAverage, setIntTempAverage] = useState<number | null>(null);
    const [tcAverage, setTCAverage] = useState<number | null>(null);
    const [logs, setLogs] = useState<Array<string>>([]);
    const [vDSRate, setVDSRate] = useState<number | null>(null);
    const [vGSRate, setVGSRate] = useState<number | null>(null);
    const [tcRate, setTCRate] = useState<number | null>(null);


    // let currentIP: string = "";
    // const [error, setError] = useState(String);

    useEffect(() => {
        // Greet("test").then((e)=>{console.log(e)});
        const fetchConnectedIPs = async () => {
            try {
                const response = await GetAllConnectedIPs();
                // console.log('Raw response:', response);
                setConnectedIPs(response || {});
                if (selectedIP && !Object.keys(response).includes(selectedIP)) {
                    setSelectedIP(null)
                }
                // setError(null);
            } catch (err) {
                console.error('Error fetching IPs:', err);
                // setError(`Failed to fetch connected IPs: ${err.message}`);
                setConnectedIPs({});
            }
        };

        fetchConnectedIPs();
        const interval = setInterval(fetchConnectedIPs, 200);
        return () => clearInterval(interval);
    }, [selectedIP]);

    useEffect(() => {
        if (!selectedIP) {
            setLogs([]);
            return
        }
        const fetchLogs = async () => {
            GetLogs(selectedIP).then((e) => {
                setLogs(e)
            }).catch((err) => {
                console.log("Error fetching logs: ", err)
            })

        };

        fetchLogs();
        const interval = setInterval(fetchLogs, 200);
        return () => clearInterval(interval);
    }, [selectedIP]);

    useEffect(() => {
        if (!selectedIP) {
            setVDSRate(null);
            setVGSRate(null);
            setTCRate(null);
            return
        }
        const getPortRates = async () => {
            let vdsKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5555
            }
            let vgsKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5556
            }
            let tcKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5557
            }
            GetPortRate(vdsKey).then((e) => {
                setVDSRate(e)
                return GetPortRate(vgsKey)
            }).then(e => {
                setVGSRate(e)
                return GetPortRate(tcKey)
            }).then(e => {
                setTCRate(e)
            })
                .catch((err) => {
                    console.log("Error fetching port rates: ", err)
                })

        };

        getPortRates();
        const interval = setInterval(getPortRates, 500);
        return () => clearInterval(interval);
    }, [selectedIP]);

    useEffect(() => {
        if (!selectedIP) {
            setLogs([]);
            return
        }
        const fetchLogs = async () => {
            GetLogs(selectedIP).then((e) => {
                setLogs(e)
            }).catch((err) => {
                console.log("Error fetching logs: ", err)
            })

        };

        fetchLogs();
        const interval = setInterval(fetchLogs, 200);
        return () => clearInterval(interval);
    }, [selectedIP]);

    useEffect(() => {
        // Set up the animation frame for GetPortAverage

        if (!selectedIP) {
            setVdsAverage(null);
            setVgsAverage(null);
            setIntTempAverage(null);
            setTCAverage(null);
            return; // Exit early if no IP is selected
        }

        // Set up the animation frame for GetPortAverage
        let animationFrameId: number;
        let isRunning = true;
        const updatePortAverage = () => {
            if (!isRunning) return;
            let vdsKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5555
            }
            let vgsKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5556
            }
            let tcKey: server.BufferKey = {
                IP: selectedIP.replace(/_/g, '.'),
                Port: 5557
            }
            // console.log(key)
            // console.log(currentIP)
            // Call GetPortAverage on every frame
            // First call GetPortAverage on vdsKey
            GetPortAverage(vdsKey)
                .then(vdsResult => {
                    setVdsAverage(vdsResult);

                    // Then call GetPortAverage on vgsKey
                    return GetPortAverage(vgsKey);
                })
                .then(vgsResult => {
                    setVgsAverage(vgsResult);
                    return GetPortAverage(tcKey);
                })
                .then(tcResult => {
                    setIntTempAverage(tcResult);
                    return GetPortAverageB(tcKey);
                })
                .then(tcResultB => {
                    setTCAverage(tcResultB)
                })
                .catch(error => {
                    console.error("Error fetching port averages:", error);
                })
                .finally(() => {
                    // Request next frame only after both calls complete
                    if (isRunning) {
                        animationFrameId = requestAnimationFrame(updatePortAverage);
                    }
                });
        };

        requestAnimationFrame(updatePortAverage);

        // Cleanup function
        return () => {
            isRunning = false;
            if (animationFrameId) {
                cancelAnimationFrame(animationFrameId);
            }
        };
    }, [selectedIP])

    const handleIPSelect = (ip: string) => {
        setSelectedIP(ip);
        // currentIP = ip;
        let devNum = Object.keys(connectedIPs).findIndex(p => p === ip);
        setSelectedDevNum(devNum)
        console.log('Selected IP:', ip);
    };


    return (
        <>
            {/* <Theme theme="g100"> */}

            <Header aria-label="Header">
                <SkipToContent />
                <HeaderName href="#" prefix="Infinus Electronics">
                    Power Cycling Control Panel
                </HeaderName>
            </Header>
            <SideNav isFixedNav expanded={true} isChildOfHeader={false} aria-label="Side navigation">
                <SideNavItems>
                    <SideNavMenu title="Connected Devices" defaultExpanded>

                        {Object.entries(connectedIPs).map(([ip, conn], _i) => {
                            return (
                                <SideNavMenuItem href="#" key={ip} isActive={selectedIP === ip} onClick={(e: React.MouseEvent) => {
                                    e.preventDefault();
                                    handleIPSelect(ip);
                                }}>
                                    {ip.replace(/_/g, '.')}
                                </SideNavMenuItem>)
                        })}

                    </SideNavMenu>

                    <SideNavDivider />
                    <SideNavLink href="#">
                        L0 link
                    </SideNavLink>
                    <SideNavLink href="#">
                        L0 link
                    </SideNavLink>
                </SideNavItems>
            </SideNav>
            {selectedIP ?
                <Content>

                    <Grid fullWidth>

                        <Column lg={16} md={8} sm={4}>
                            <h1>
                                Device {selectedDevNum}
                            </h1>
                        </Column>
                    </Grid>
                    <Grid fullWidth>
                        <Column lg={16} md={8} sm={4}>
                            <h2 className="inf-section-heading">Device Info</h2>
                        </Column>

                        <Column lg={16} md={8} sm={4}>
                            <p className='inf-device-info'>
                                UUID:
                            </p>
                            <p className='inf-device-info-value'>
                                {connectedIPs[selectedIP] ? connectedIPs[selectedIP].UUID:"N/A"}
                            </p>
                        </Column>

                        <Column lg={4} md={4} sm={2}>
                            <p className='inf-device-info'>
                                IP Address:
                            </p>
                            <p className='inf-device-info-value'>
                                {selectedIP ? selectedIP.replace(/_/g, '.') : 'N/A'}
                            </p>
                        </Column>
                        <Column lg={4} md={4} sm={2}>
                            <p className='inf-device-info'>
                                V<sub>DS</sub> Port Data Rate:
                            </p>
                            <p className='inf-device-info-value'>
                                {vDSRate ? vDSRate.toFixed(3) : "N/A"} MB/s
                            </p>
                        </Column>

                        <Column lg={4} md={4} sm={2}>
                            <p className='inf-device-info'>
                                V<sub>GS</sub> Port Data Rate:
                            </p>
                            <p className='inf-device-info-value'>
                                {vGSRate ? vGSRate.toFixed(3) : "N/A"} MB/s
                            </p>
                        </Column>
                        <Column lg={4} md={4} sm={2}>
                            <p className='inf-device-info'>
                                Thermocouple Port Data Rate:
                            </p>
                            <p className='inf-device-info-value'>
                                {tcRate ? (tcRate * 1000).toFixed(3) : "N/A"} kB/s
                            </p>
                        </Column>

                    </Grid>


                    <Grid fullWidth>
                        <Column lg={16} md={8} sm={4}>
                            <h2 className="inf-section-heading">
                                Live Measurement Results
                            </h2>
                        </Column>
                        <Column lg={8} md={8} sm={4}>
                        <Grid>
                            <Column span={16}>
                            <h4>
                                V<sub>DS</sub>: <span className='inf-meas-result-value'>{
                                    vdsAverage === null || vdsAverage === undefined
                                        ? "-.-"
                                        : (vdsAverage < 0 ? vdsAverage.toFixed(5) : vdsAverage.toFixed(3))
                                } V </span>
                            </h4>
                            </Column>
                            <Column span = {16}>
                            <h4>
                                V<sub>GS</sub>: <span className='inf-meas-result-value'>{
                                    vgsAverage === null || vgsAverage === undefined
                                        ? "-.-"
                                        : vgsAverage.toFixed(4)
                                } V </span>
                            </h4>
                        </Column>
                        </Grid>
                            
                        </Column>

                        

                        <Column lg={8} md={8} sm={4}>
                        <Grid>
                        <Column span = {16}>
                        <h4>
                                Internal Temperature: <span className='inf-meas-result-value'>{
                                    intTempAverage === null || intTempAverage === undefined
                                        ? "-.-"
                                        : intTempAverage.toFixed(4)
                                } °C</span>
                            </h4>
                        </Column>
                        <Column span={16}>
                            <h4>
                                Thermocouple Temperature: <span className='inf-meas-result-value'>{
                                    tcAverage === null || tcAverage === undefined
                                        ? "-.-"
                                        : tcAverage.toFixed(4)
                                } °C</span>
                            </h4>
                        </Column>
                        </Grid>
                            
                        </Column>

                        
                        </Grid>
                        <Grid fullWidth>
                        <Column lg={16} md={8} sm={4}>
                            <h2 className="inf-section-heading">
                                Device Logs
                            </h2>
                        </Column>

                        <Column lg={16} md={8} sm={4}>
                            <div className='inf-device-logs-container'>
                                {logs.map((e, i) => {
                                    return (
                                        <p key={i} className='inf-device-log-entry'>
                                            {e}
                                        </p>
                                    )
                                })}
                            </div>
                        </Column>
                    </Grid>


                </Content>
                : <></>}

            {/* </Theme> */}
        </>


    )
}

export default App
