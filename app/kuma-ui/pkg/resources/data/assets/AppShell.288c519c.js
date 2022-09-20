import{m as x,H as P,q as F,_ as b,r,o,f as a,d as p,w as h,u as A,n as k,h as s,t as v,p as C,b as m,e as g,F as $,j as I,I as B,z as K,B as L,k as y,J as D,L as V,x as R,M as W,N as O,O as z,G as H,T as j}from"./index.6180ff6f.js";import{d as G}from"./datadogEvents.f0c2b8e0.js";import{A as q,a as Z}from"./AccordionItem.ddeeebdc.js";const U={name:"NavItem",props:{link:{type:String,default:"",required:!1},insightsFieldAccessor:{type:String,default:"",required:!1},name:{type:String,default:""},icon:{type:String,default:""},hasIcon:{type:Boolean,default:!1},hasCustomIcon:{type:Boolean,default:!1},isMenuItem:{type:Boolean,default:!0},isDisabled:{type:Boolean,default:!1},title:{type:Boolean,default:!1},nested:{type:Boolean,default:!1},usesMeshParam:{type:Boolean,required:!1,default:!1}},data(){return{meshPath:null}},computed:{...x({selectedMesh:e=>e.selectedMesh,insights:e=>e.sidebar.insights}),insightsClassess(){return["amount",{"amount--empty":this.amount===0}]},amount(){const e=P(this.insights,this.insightsFieldAccessor,0);return e>99?"99+":e},routerLink(){const e={};return this.link?e.name=this.link:this.title?e.name=null:e.name=this.$route.name,this.usesMeshParam&&(e.params={mesh:this.selectedMesh}),e},isActive(){const e=this.link,n=this.$route,t=this.$route.path.split("/")[2];return e===n.name||t===this.routerLink.name?!0:e&&n.matched.some(d=>e===d.name||e===d.redirect)}},methods:{onNavItemClick(){F.logger.info(G.SIDEBAR_ITEM_CLICKED,{data:this.routerLink})}}},J=["data-testid"],Y={key:0,class:"nav-icon"},Q={key:1,class:"title-text"},X={class:"text-uppercase"},ee={key:2,class:"nav-link"};function te(e,n,t,d,c,i){const l=r("KIcon"),u=r("router-link");return o(),a("div",{class:C([[{"is-active":i.isActive},{"is-menu-item":t.isMenuItem},{"is-disabled":t.isDisabled},{"is-title":t.title},{"is-nested":t.nested}],"nav-item"]),"data-testid":t.link},[p(u,{to:i.routerLink,onClick:i.onNavItemClick},{default:h(()=>[t.hasIcon||t.hasCustomIcon?(o(),a("div",Y,[A(e.$slots,"item-icon",{},()=>[t.hasIcon&&t.icon?(o(),m(l,{key:0,width:"18",height:"18",color:"var(--SidebarIconColor)",icon:t.icon},null,8,["icon"])):k("",!0)],!0)])):k("",!0),t.title?(o(),a("div",Q,[s("span",X,v(t.name),1)])):(o(),a("div",ee,[A(e.$slots,"item-link",{},()=>[g(v(t.name)+" ",1),t.insightsFieldAccessor?(o(),a("span",{key:0,class:C(i.insightsClassess)},v(i.amount),3)):k("",!0)],!0)]))]),_:3},8,["to","onClick"])],10,J)}const E=b(U,[["render",te],["__scopeId","data-v-41655ff2"]]);const se={name:"SubNav",components:{NavItem:E},props:{title:{type:String,default:""},items:{type:Array,required:!0},titleLink:{type:String,default:""}},emits:["toggled"],data(){return{isCollapsed:!1}},computed:{touchDevice(){return!!("ontouchstart"in window||navigator.maxTouchPoints)}},methods:{handleToggle(){this.touchDevice&&(this.isCollapsed=!this.isCollapsed,this.$emit("toggled",this.isCollapsed))}}},ne={class:"mt-3"},oe={class:"subnav-title"},ie={class:"text-uppercase"};function ae(e,n,t,d,c,i){const l=r("router-link"),u=r("NavItem");return o(),a("div",{class:C([{"is-collapsed":c.isCollapsed},"secondary-nav"])},[s("div",ne,[A(e.$slots,"top",{},void 0,!0)]),s("div",oe,[s("span",ie,[A(e.$slots,"title",{},()=>[p(l,{to:{name:t.titleLink}},{default:h(()=>[g(v(t.title),1)]),_:1},8,["to"])],!0)])]),A(e.$slots,"bottom",{},void 0,!0),(o(!0),a($,null,I(t.items,(M,f)=>(o(),m(u,B({key:f},M),null,16))),128))],2)}const ce=b(se,[["render",ae],["__scopeId","data-v-921799f1"]]);const re={name:"MeshSelector",props:{items:{type:Object,required:!0}},computed:{...x({selectedMesh:e=>e.selectedMesh})},methods:{changeMesh(e){const n=e.target.value;this.$store.dispatch("updateSelectedMesh",n),this.$router.push({name:this.$route.name,params:"mesh"in this.$route.params?{mesh:n}:void 0})}}},le=e=>(K("data-v-d27bb047"),e=e(),L(),e),de={class:"px-4 pb-4"},ue={key:0},he=le(()=>s("h3",{class:"menu-title uppercase"}," Filter by Mesh: ",-1)),me=["selected"],_e=["value","selected"];function pe(e,n,t,d,c,i){const l=r("KAlert");return o(),a("div",de,[t.items?(o(),a("div",ue,[he,s("select",{id:"mesh-selector",class:"mesh-selector",name:"mesh-selector",onChange:n[0]||(n[0]=(...u)=>i.changeMesh&&i.changeMesh(...u))},[s("option",{value:"all",selected:e.selectedMesh==="all"}," All Meshes ",8,me),(o(!0),a($,null,I(t.items.items,u=>(o(),a("option",{key:u.name,value:u.name,selected:u.name===e.selectedMesh},v(u.name),9,_e))),128))],32)])):(o(),m(l,{key:1,appearance:"danger","alert-message":"No meshes found!"}))])}const fe=b(re,[["render",pe],["__scopeId","data-v-d27bb047"]]),ge=`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 21 19"><g fill="none" fill-rule="evenodd" transform="translate(1)"><path stroke="#1456CB" stroke-opacity=".4" stroke-width="1.5" d="M2.75 2.75h13.5v13.5H2.75zM3.5 3.5l12 12M3.5 15.5l12-12"/><circle cx="2.5" cy="2.5" r="2.5" fill="#1456CB" fill-rule="nonzero"/><circle cx="16.5" cy="2.5" r="2.5" fill="#1456CB" fill-rule="nonzero"/><circle cx="9.5" cy="9.5" r="2.5" fill="#1456CB" fill-rule="nonzero"/><circle cx="2.5" cy="16.5" r="2.5" fill="#1456CB" fill-rule="nonzero"/><circle cx="16.5" cy="16.5" r="2.5" fill="#1456CB" fill-rule="nonzero"/></g></svg>
`;function ve(e){const n=e.map(t=>({name:t.pluralDisplayName,link:t.path,title:!1,usesMeshParam:!0,parent:"policies",insightsFieldAccessor:`mesh.policies.${t.name}`}));return n.sort((t,d)=>t.name<d.name?-1:1),[{name:"Service Mesh",iconCustom:ge,link:"home",subNav:{items:[{name:"Overview",link:"global-overview",usesMeshParam:!0},{name:"Meshes",link:"mesh-child",pathFlip:!0,usesMeshParam:!0,insightsFieldAccessor:"global.Mesh"},{name:"Zones",title:!0},{name:"Zone CPs",link:"zones",insightsFieldAccessor:"global.Zone"},{name:"Zone Ingresses",link:"zoneingresses",insightsFieldAccessor:"global.ZoneIngress"},{name:"Zone Egresses",link:"zoneegresses",insightsFieldAccessor:"global.ZoneEgress"},{name:"Services",title:!0},{name:"Internal",link:"internal-services",title:!1,usesMeshParam:!0,insightsFieldAccessor:"mesh.services.internal"},{name:"External",link:"external-services",title:!1,usesMeshParam:!0,insightsFieldAccessor:"mesh.services.external"},{name:"Data plane proxies",title:!0},{name:"All",link:"dataplanes",title:!1,usesMeshParam:!0,insightsFieldAccessor:"mesh.dataplanes.total"},{name:"Standard",link:"standard-dataplanes",title:!1,nested:!0,usesMeshParam:!0,insightsFieldAccessor:"mesh.dataplanes.standard"},{name:"Gateway",link:"gateway-dataplanes",title:!1,nested:!0,usesMeshParam:!0,insightsFieldAccessor:"mesh.dataplanes.gateway"},{name:"Policies",title:!0},...n]}}]}function be(){return[{name:"Diagnostics",icon:"gearFilled",link:"diagnostics"}]}const Me={name:"AppSidebar",components:{MeshSelector:fe,NavItem:E,Subnav:ce},data(){return{isCollapsed:!1,sidebarSavedState:null,toggleWorkspaces:!1,isHovering:!1,subnavIsExpanded:!0,topMenuItems:[],bottomMenuItems:be()}},computed:{...x({selectedMesh:e=>e.selectedMesh,policies:e=>e.policies}),...y({featureFlags:"config/featureFlags"}),topNavItems(){return this.topMenuItems.length>0?this.topMenuItems[0].subNav.items.filter(e=>e.featureFlags?e.featureFlags.every(n=>this.featureFlags.includes(n)):!0):[]},hasSubnav(){var e,n,t;return Boolean((t=(n=(e=this.selectedMenuItem)==null?void 0:e.subNav)==null?void 0:n.items)==null?void 0:t.length)},lastMenuList(){return Object.keys(this.menuList.sections).length-1},meshList(){return this.$store.state.meshes},selectedMenuItem(){if(this.topMenuItems.length===0)return null;const e=this.$route;for(const n of[...this.topMenuItems,...this.bottomMenuItems])for(const t of n.items)if(e.name!==t.link&&!e.meta.hideSubnav)return t;return null},touchDevice(){return!!("ontouchstart"in window||navigator.maxTouchPoints)}},watch:{selectedMesh(e){this.getMeshInsights()}},created(){this.topMenuItems=ve(this.policies)},mounted(){this.sidebarEvent()},beforeUnmount(){},methods:{...D({getMeshInsights:"sidebar/getMeshInsights"}),handleResize(){const e=V.innerWidth;e<=900&&(this.isCollapsed=!0,this.subnavIsExpanded=!1,this.isHovering=!1),e>=900&&(this.isCollapsed=!1,this.isHovering=!0)},toggleSubnav(){this.subnavIsExpanded=!0,this.isCollapsed=!0,localStorage.setItem("sidebarCollapsed",this.subnavIsExpanded)},sidebarEvent(){const e=this.touchDevice,n=this.$refs.sidebarControl;this.$route.params.expandSidebar&&this.$route.params.expandSidebar===!0&&(this.subnavIsExpanded=!0,localStorage.setItem("sidebarCollapsed",!0)),e?(n.addEventListener("touchstart",()=>{this.isHovering=!0}),n.addEventListener("touchend",()=>{this.isHovering=!1})):(n.addEventListener("mouseover",()=>{this.isHovering=!0}),n.addEventListener("mouseout",()=>{this.isHovering=!1}),n.addEventListener("click",()=>{this.isHovering=!1}))}}},ye={class:"top-nav"},ke=["innerHTML"],$e={class:"bottom-nav"};function Ie(e,n,t,d,c,i){const l=r("NavItem"),u=r("MeshSelector"),M=r("Subnav");return o(),a("aside",{id:"the-sidebar",class:C(["has-subnav",[{"is-collapsed":c.isCollapsed},{"subnav-expanded":c.subnavIsExpanded}]])},[s("div",{ref:"sidebarControl",class:C(["main-nav",{"is-hovering":c.isHovering||c.subnavIsExpanded===!1}])},[s("div",ye,[(o(!0),a($,null,I(c.topMenuItems,(f,_)=>(o(),m(l,B({key:_},f,{"has-custom-icon":"",onClick:i.toggleSubnav}),R({_:2},[f.iconCustom&&!f.icon?{name:"item-icon",fn:h(()=>[s("div",{innerHTML:f.iconCustom},null,8,ke)]),key:"0"}:void 0]),1040,["onClick"]))),128))]),s("div",$e,[(o(!0),a($,null,I(c.bottomMenuItems,(f,_)=>(o(),m(l,B({key:_},f,{"has-icon":""}),null,16))),128))])],2),c.subnavIsExpanded&&c.topMenuItems.length>0?(o(),m(M,{key:0,title:c.topMenuItems[0].name,"title-link":c.topMenuItems[0].link,items:i.topNavItems},{top:h(()=>[p(u,{items:i.meshList},null,8,["items"])]),_:1},8,["title","title-link","items"])):k("",!0)],2)}const Ae=b(Me,[["render",Ie]]);const Ce={name:"AllMeshesNotifications",emits:["mesh-selected"],data(){return{url:"https://kuma.io/enterprise/?utm_source=Kuma&utm_medium=Kuma-GUI"}},computed:{...y({meshNotificationItemMapWithAction:"notifications/meshNotificationItemMapWithAction"}),hasMeshesWithAction(){return Object.keys(this.meshNotificationItemMapWithAction).length>0}},methods:{meshSelected(e){this.$emit("mesh-selected",e)},calculateActions(e){const n=Object.values(e),t=n.filter(Boolean);return n.length-t.length}}},w=e=>(K("data-v-599d512f"),e=e(),L(),e),Ne={class:"py-4"},Se=w(()=>s("h3",{class:"font-bold mb-4"}," Meshes ",-1)),xe={key:0},we=w(()=>s("p",null," Check the following meshes for suggestions to adjust the configuration ",-1)),Be={class:"pt-4 flex space-x-4"},Ke={class:"notification-amount"},Le={key:1},De={class:"py-4"},Oe=w(()=>s("h3",{class:"font-bold mb-4"}," Enterprise ",-1)),Ee=w(()=>s("p",null," Kuma\u2019s ecosystem has created enterprise offerings to do more with the product, including advanced integrations and support. ",-1)),Te=g(" Kuma Enterprise Offerings ");function Pe(e,n,t,d,c,i){const l=r("KBadge"),u=r("KIcon"),M=r("KButton");return o(),a("div",null,[s("div",Ne,[Se,i.hasMeshesWithAction?(o(),a("div",xe,[we,s("div",Be,[(o(!0),a($,null,I(e.meshNotificationItemMapWithAction,(f,_)=>(o(),a("span",{key:_,class:"relative d-inline-block"},[p(l,{class:"cursor-pointer hover:scale-110",onClick:N=>i.meshSelected(_)},{default:h(()=>[g(v(_),1)]),_:2},1032,["onClick"]),s("span",Ke,v(i.calculateActions(f)),1)]))),128))])])):(o(),a("div",Le," Looks like none of your meshes are missing any features. Well done! "))]),s("div",De,[Oe,Ee,p(M,{class:"enterprise-button",appearance:"primary",target:"_blank",to:c.url},{default:h(()=>[p(u,{icon:"organizations",color:"white",size:"24"}),Te]),_:1},8,["to"])])])}const Fe=b(Ce,[["render",Pe],["__scopeId","data-v-599d512f"]]);const Ve={name:"OnboardingNotification",data(){return{alertClosed:!1,productName:W}},methods:{closeAlert(){this.alertClosed=!0}}},Re={key:0,class:"onboarding-check"},We={class:"alert-content"},ze=g(" We've detected that you don't have any data plane proxies running yet. We've created an onboarding process to help you! "),He=g(" Get Started ");function je(e,n,t,d,c,i){const l=r("KButton"),u=r("KAlert");return c.alertClosed===!1?(o(),a("div",Re,[p(u,{appearance:"success",class:"dismissible","dismiss-type":"icon",onClosed:i.closeAlert},{alertMessage:h(()=>[s("div",We,[s("div",null,[s("strong",null,"Welcome to "+v(c.productName)+"!",1),ze]),s("div",null,[p(l,{appearance:"primary",size:"small",class:"action-button",to:{name:"onboarding-welcome"}},{default:h(()=>[He]),_:1})])])]),_:1},8,["onClosed"])])):k("",!0)}const T=b(Ve,[["render",je],["__scopeId","data-v-b6c3e5d4"]]),Ge={name:"LoggingNotification",computed:{...y({kumaDocsVersion:"config/getKumaDocsVersion"})}},qe={class:"py-4"},Ze=s("p",{class:"mb-4"}," A traffic log policy lets you collect access logs for every data plane proxy in your service mesh. ",-1),Ue={class:"list-disc pl-4"},Je=["href"];function Ye(e,n,t,d,c,i){return o(),a("div",qe,[Ze,s("ul",Ue,[s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/traffic-log/`,target:"_blank"}," Traffic Log policy documentation ",8,Je)])])])}const Qe=b(Ge,[["render",Ye]]),Xe={name:"MetricsNotification",computed:{...y({kumaDocsVersion:"config/getKumaDocsVersion"})}},et={class:"py-4"},tt=s("p",{class:"mb-4"}," A traffic metrics policy lets you collect key data for observability of your service mesh. ",-1),st={class:"list-disc pl-4"},nt=["href"];function ot(e,n,t,d,c,i){return o(),a("div",et,[tt,s("ul",st,[s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/traffic-metrics/`,target:"_blank"}," Traffic Metrics policy documentation ",8,nt)])])])}const it=b(Xe,[["render",ot]]),at={name:"MtlsNotification",computed:{...y({kumaDocsVersion:"config/getKumaDocsVersion"})}},ct={class:"py-4"},rt=s("p",{class:"mb-4"}," Mutual TLS (mTLS) for communication between all the components of your service mesh (services, control plane, data plane proxies), proxy authentication, and access control rules in Traffic Permissions policies all contribute to securing your mesh. ",-1),lt={class:"list-disc pl-4"},dt=["href"],ut=["href"],ht=["href"];function mt(e,n,t,d,c,i){return o(),a("div",ct,[rt,s("ul",lt,[s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/security/certificates/`,target:"_blank"}," Secure access across services ",8,dt)]),s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/mutual-tls/`,target:"_blank"}," Mutual TLS ",8,ut)]),s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/traffic-permissions/`,target:"_blank"}," Traffic Permissions policy documentation ",8,ht)])])])}const _t=b(at,[["render",mt]]),pt={name:"TracingNotification",computed:{...y({kumaDocsVersion:"config/getKumaDocsVersion"})}},ft={class:"py-4"},gt=s("p",{class:"mb-4"}," A traffic trace policy lets you enable tracing logs and a third-party tracing solution to send them to. ",-1),vt={class:"list-disc pl-4"},bt=["href"];function Mt(e,n,t,d,c,i){return o(),a("div",ft,[gt,s("ul",vt,[s("li",null,[s("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/traffic-trace/`,target:"_blank"}," Traffic Trace policy documentation ",8,bt)])])])}const yt=b(pt,[["render",Mt]]),kt={name:"SingleMeshNotifications",components:{AccordionList:q,AccordionItem:Z,OnboardingNotification:T,LoggingNotification:Qe,MetricsNotification:it,MtlsNotification:_t,TracingNotification:yt},computed:{...y({singleMeshNotificationItems:"notifications/singleMeshNotificationItems"})}},$t={class:"flex items-center"};function It(e,n,t,d,c,i){const l=r("KIcon"),u=r("KCard"),M=r("AccordionItem"),f=r("AccordionList");return o(),m(f,{"multiple-open":""},{default:h(()=>[(o(!0),a($,null,I(e.singleMeshNotificationItems,_=>(o(),m(M,{key:_.name},{"accordion-header":h(()=>[s("div",$t,[_.isCompleted?(o(),m(l,{key:0,color:"var(--green-400)",icon:"check",size:"20",class:"mr-4"})):(o(),m(l,{key:1,icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20",class:"mr-4"})),s("strong",null,v(_.name),1)])]),"accordion-content":h(()=>[_.component?(o(),m(O(_.component),{key:0})):(o(),m(u,{key:1},{body:h(()=>[g(v(_.content),1)]),_:2},1024))]),_:2},1024))),128))]),_:1})}const At=b(kt,[["render",It]]);const Ct={name:"NotificationManager",components:{AllMeshesNotifications:Fe,SingleMeshNotifications:At},data(){return{alertClosed:!1}},computed:{...x({isOpen:e=>e.notifications.isOpen,selectedMesh:e=>e.selectedMesh}),...y({amountOfActions:"notifications/amountOfActions",showOnboarding:"onboarding/showOnboarding",meshNotificationItemMapWithAction:"notifications/meshNotificationItemMapWithAction"}),isAllMeshesView(){return this.selectedMesh==="all"},shouldRenderAlert(){return!this.alertClosed&&!this.showOnboarding&&this.amountOfActions>0},hasAnyAction(){return this.meshNotificationItemMapWithAction[this.selectedMesh]}},methods:{...D({openModal:"notifications/openModal",closeModal:"notifications/closeModal",updateSelectedMesh:"updateSelectedMesh"}),closeAlert(){this.alertClosed=!0},changeMesh(e){this.updateSelectedMesh(e),this.$router.push({name:this.$route.name,params:{mesh:e}})}}},Nt=e=>(K("data-v-4fad15b5"),e=e(),L(),e),St={class:"mr-4"},xt={class:"mr-2"},wt=Nt(()=>s("strong",null,"ProTip:",-1)),Bt={key:0,class:"flex items-center"},Kt=g(" Notifications "),Lt={key:1},Dt={key:0},Ot=g(" Some of these features are not enabled for "),Et={class:"text-xl tracking-wide"},Tt=g(" mesh. Consider implementing them. "),Pt={key:1},Ft=g(" Looks like "),Vt={class:"text-xl tracking-wide"},Rt=g(" isn't missing any features. Well done! "),Wt=g(" \u2039 Back to all "),zt=g(" Close ");function Ht(e,n,t,d,c,i){const l=r("KButton"),u=r("KAlert"),M=r("KIcon"),f=r("AllMeshesNotifications"),_=r("SingleMeshNotifications"),N=r("KModal");return o(),a("div",null,[i.shouldRenderAlert?(o(),m(u,{key:0,class:"mb-4",appearance:"info","dismiss-type":"icon","data-testid":"notification-info",onClosed:i.closeAlert},{alertMessage:h(()=>[s("div",St,[s("span",xt,[wt,g(" You might want to adjust your "+v(i.isAllMeshesView?"meshes":"mesh")+" configuration ",1)]),p(l,{appearance:"outline",onClick:e.openModal},{default:h(()=>[g(" Check your "+v(i.isAllMeshesView?"meshes":"mesh")+"! ",1)]),_:1},8,["onClick"])])]),_:1},8,["onClosed"])):k("",!0),p(N,{class:"modal","is-visible":e.isOpen,title:"Notifications","text-align":"left"},{"header-content":h(()=>[i.isAllMeshesView?(o(),a("div",Bt,[p(M,{color:"var(--yellow-300)",icon:"notificationBell",size:"24",class:"mr-2"}),Kt])):(o(),a("div",Lt,[s("div",null,[i.hasAnyAction?(o(),a("span",Dt,[Ot,s("span",Et,' "'+v(e.selectedMesh)+'"',1),Tt])):(o(),a("span",Pt,[Ft,s("span",Vt,' "'+v(e.selectedMesh)+'"',1),Rt]))]),p(l,{class:"mt-4",appearance:"outline",onClick:n[0]||(n[0]=S=>i.changeMesh("all"))},{default:h(()=>[Wt]),_:1})]))]),"body-content":h(()=>[i.isAllMeshesView?(o(),m(f,{key:0,onMeshSelected:n[1]||(n[1]=S=>i.changeMesh(S))})):(o(),m(_,{key:1}))]),"footer-content":h(()=>[p(l,{appearance:"outline",onClick:e.closeModal},{default:h(()=>[zt]),_:1},8,["onClick"])]),_:1},8,["is-visible"])])}const jt=b(Ct,[["render",Ht],["__scopeId","data-v-4fad15b5"]]);const Gt={name:"BreadcrumbsMenu",components:{KBreadcrumbs:z},computed:{pageMesh(){return this.$route.params.mesh},routes(){const e=[];this.$route.matched.forEach(t=>{const d=t.redirect!==void 0&&t.redirect.name!==void 0?t.redirect.name:t.name;this.isCurrentRoute(t)&&this.pageMesh&&e.push({key:this.pageMesh,to:{path:`/meshes/${this.pageMesh}`},title:`Mesh Overview for ${this.pageMesh}`,text:this.pageMesh}),this.isCurrentRoute(t)&&t.meta.parent&&t.meta.parent!=="undefined"?e.push({key:t.meta.parent,to:{name:t.meta.parent},title:t.meta.title,text:t.meta.breadcrumb||t.meta.title}):this.isCurrentRoute(t)&&!t.meta.excludeAsBreadcrumb?e.push({key:d,to:{name:d},title:t.meta.title,text:t.meta.breadcrumb||t.meta.title}):t.meta.parent&&t.meta.parent!=="undefined"&&e.push({key:t.meta.parent,to:{name:t.meta.parent},title:t.meta.title,text:t.meta.breadcrumb||t.meta.title})});const n=this.calculateRouteTextAdvanced(this.$route);return n&&e.push({title:n,text:n}),e},hideBreadcrumbs(){return this.$route.query.hide_breadcrumb}},methods:{isCurrentRoute(e){return e.name&&e.name===this.$route.name||e.redirect===this.$route.name},calculateRouteTextAdvanced(e){const n=e.params,{expandSidebar:t,...d}=n,c=e.name==="mesh-overview",i=Object.assign({},d,{mesh:null});return c?n.mesh:Object.values(i).filter(l=>l)[0]}}};function qt(e,n,t,d,c,i){const l=r("KBreadcrumbs");return i.routes.length>0&&!i.hideBreadcrumbs?(o(),m(l,{key:0,items:i.routes},null,8,["items"])):k("",!0)}const Zt=b(Gt,[["render",qt]]),Ut={name:"AppShell",components:{BreadcrumbsMenu:Zt,AppSidebar:Ae,NotificationManager:jt,OnboardingNotification:T,GlobalHeader:H},computed:{...y({showOnboarding:"onboarding/showOnboarding"}),routeKey(){return this.$route.meta.shouldReRender?this.$route.path:"default"}}},Jt={class:"main-content-container"},Yt={class:"main-content"};function Qt(e,n,t,d,c,i){const l=r("GlobalHeader"),u=r("AppSidebar"),M=r("NotificationManager"),f=r("OnboardingNotification"),_=r("BreadcrumbsMenu"),N=r("router-view");return o(),a("div",null,[p(l),s("div",Jt,[p(u),s("main",Yt,[p(M),e.showOnboarding?(o(),m(f,{key:0})):k("",!0),p(_),(o(),m(N,{key:i.routeKey},{default:h(({Component:S})=>[p(j,{mode:"out-in",name:"fade"},{default:h(()=>[(o(),m(O(S)))]),_:2},1024)]),_:1}))])])])}const ss=b(Ut,[["render",Qt]]);export{ss as default};
