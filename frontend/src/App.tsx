import { useEffect, useState } from 'react';
import { Greet, GetAllConnectedIPs } from "../wailsjs/go/main/App";
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
    const [selectedDevNum, setSelectedDevNum] = useState<number|null>(null);
    // const [error, setError] = useState(String);

    useEffect(() => {
        // Greet("test").then((e)=>{console.log(e)});
        const fetchConnectedIPs = async () => {
            try {
                const response = await GetAllConnectedIPs();
                console.log('Raw response:', response);
                setConnectedIPs(response || {});
                // setError(null);
            } catch (err) {
                console.error('Error fetching IPs:', err);
                // setError(`Failed to fetch connected IPs: ${err.message}`);
                setConnectedIPs({});
            }
        };

        fetchConnectedIPs();
        const interval = setInterval(fetchConnectedIPs, 1000);
        return () => clearInterval(interval);
    }, []);

    const handleIPSelect = (ip: string) => {
        setSelectedIP(ip);
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
            <Content>
                <Grid fullWidth>

                    <Column lg={16} md={8} sm={4}>
                        <h1>
                            Device {selectedDevNum}
                        </h1>
                    </Column>
                    </Grid>
                        <Grid fullWidth>

                            <Column lg={8} md={4} sm={2}>
                                <p className='inf-device-info'>
                                    MAC Address:
                                </p>
                                <p className='inf-device-info-value'>
                                    FF:FF:FF:FF:FF:FF
                                </p>
                            </Column>

                            <Column lg={8} md={4} sm={2}>
                                <p className='inf-device-info'>
                                    IP Address:
                                </p>
                                <p className='inf-device-info-value'>
                                    {selectedIP ? selectedIP.replace(/_/g, '.') : 'N/A'}
                                </p>
                            </Column>

                        </Grid>



                <Grid fullWidth>
                    <Column lg={16} md={8} sm={4}>
                        <h2>
                            V<sub>DS</sub>
                        </h2>
                    </Column>
                </Grid>


            </Content>


            {/* </Theme> */}
        </>


    )
}

export default App
