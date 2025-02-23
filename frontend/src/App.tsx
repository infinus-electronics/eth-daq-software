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
                    <SideNavMenu title="Connected Devices">

                        {Object.entries(connectedIPs).map(([ip, conn], _i) => {
                            return(
                            <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                                 {ip.replace(/_/g, '.')}
                             </SideNavMenuItem>)
                            })}

                        
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                    </SideNavMenu>
                    <SideNavMenu title="L0 menu" isActive={true}>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem aria-current="page" href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                    </SideNavMenu>
                    <SideNavMenu title="L0 menu">
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                        <SideNavMenuItem href="https://www.carbondesignsystem.com/">
                            L0 menu item
                        </SideNavMenuItem>
                    </SideNavMenu>
                    <SideNavDivider />
                    <SideNavLink href="https://www.carbondesignsystem.com/">
                        L0 link
                    </SideNavLink>
                    <SideNavLink href="https://www.carbondesignsystem.com/">
                        L0 link
                    </SideNavLink>
                </SideNavItems>
            </SideNav>
            <Content>
                <Grid fullWidth>
                    <Column lg={12} md={6} sm={3}>
                        <h1>
                            Test
                        </h1>
                    </Column>
                    <Column lg={4} md={2} sm={1}>
                        <Grid>
                            <Column span={16}>
                                <p className='inf-device-info'>
                                    Device ONLINE
                                </p>
                            </Column>
                            <Column span={16}>
                                <p className='inf-device-info'>
                                    MAC Address:
                                </p>
                                <p className='inf-device-info-value'>
                                    FF:FF:FF:FF:FF:FF
                                </p>
                            </Column>
                            <Column span={16}>
                                <p className='inf-device-info'>
                                    IP Address:
                                </p>
                                <p className='inf-device-info-value'>
                                    255.255.255.255
                                </p>
                            </Column>
                        </Grid>

                    </Column>
                </Grid>
            </Content>


            {/* </Theme> */}
        </>


    )
}

export default App
