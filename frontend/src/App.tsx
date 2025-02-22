import { useState } from 'react';
import { Greet } from "../wailsjs/go/main/App";
import "./globals.scss";
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
    HeaderSideNavItems,
    Content,
    Grid,
    Column
} from '@carbon/react';
import { Switcher, Notification, UserAvatar } from '@carbon/icons-react';

function App() {
    const [resultText, setResultText] = useState("Please enter your name below 👇");
    const [name, setName] = useState('');
    const updateName = (e: any) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);

    function greet() {
        Greet(name).then(updateResultText);
    }

    return (
        <>
            <Header>
                <HeaderGlobalBar>
                    <HeaderGlobalAction
                        aria-label="Notifications"
                        tooltipAlignment="center"
                        className="action-icons">
                        <Notification size={20} />
                    </HeaderGlobalAction>
                    <HeaderGlobalAction
                        aria-label="User Avatar"
                        tooltipAlignment="center"
                        className="action-icons">
                        <UserAvatar size={20} />
                    </HeaderGlobalAction>
                    <HeaderGlobalAction aria-label="App Switcher" tooltipAlignment="end">
                        <Switcher size={20} />
                    </HeaderGlobalAction>
                </HeaderGlobalBar>
            </Header>

            <Content>
                <Grid className="landing-page" fullWidth>
                    <Column lg={16} md={8} sm={4} className="landing-page__banner">
                        1
                    </Column>
                    <Column lg={16} md={8} sm={4} className="landing-page__r2">
                        <Grid className="tabs-group-content">
                            <Column md={4} lg={7} sm={4} className="landing-page__tab-content">
                                7/16
                            </Column>
                            <Column md={4} lg={{ span: 8, offset: 8 }} sm={4}>
                                8/16
                            </Column>
                        </Grid>
                    </Column>
                    <Column lg={16} md={8} sm={4} className="landing-page__r3">
                        <Grid>
                            <Column md={4} lg={4} sm={4}>
                                1/4
                            </Column>
                            <Column md={4} lg={4} sm={4}>
                                1/4
                            </Column>
                            <Column md={4} lg={4} sm={4}>
                                1/4
                            </Column>
                            <Column md={4} lg={4} sm={4}>
                                1/4
                            </Column>
                        </Grid>
                    </Column>
                </Grid>
            </Content>
        </>

    )
}

export default App
