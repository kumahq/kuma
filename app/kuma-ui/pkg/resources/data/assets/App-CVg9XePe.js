import{d as w,r as m,o as c,c as S,a as _,b as s,w as n,e,t as f,n as O,_ as M,h as V,f as C,g as X,i as I,u as z,j as T,k as U,l as D,m as o,p as a,q as b,s as h,v as L,x as B}from"./index-CKQWVGYP.js";const x=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,K={class:"app-navigator"},G=w({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(d){const i=d;return(u,p)=>{const r=m("XAction");return c(),S("li",K,[_(u.$slots,"default",{},()=>[s(r,{class:O({"is-active":i.active}),to:i.to},{default:n(()=>[e(f(i.label),1)]),_:1},8,["class","to"])],!0)])}}}),$=M(G,[["__scopeId","data-v-07bb7885"]]),P=w({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const d={ref:"_"};for(const i in this.$props)d[V(i)]=this.$props[i];return C("span",[X(this.$slots,"default")?C("a",d,this.$slots.default()):C("a",d)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){if(this.$el.lastChild!==this.$refs._)return;const d=this.$el.appendChild(document.createElement("span")),i=this;I(()=>import("./buttons.esm-DK2fWHEW.js"),[],import.meta.url).then(function(u){i.$el.lastChild===d&&u.render(d.appendChild(i.$refs._),function(p){i.$el.lastChild===d&&d.parentNode.replaceChild(p,d)})})},reset:function(){this.$refs._!=null&&this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),H={class:"application-shell"},Y={role:"banner"},q={class:"horizontal-list"},Z={class:"upgrade-check-wrapper"},j={class:"alert-content"},F={class:"horizontal-list"},J={class:"app-status app-status--mobile"},Q={class:"app-status app-status--desktop"},W={class:"app-content-container"},tt={class:"app-sidebar"},et={"aria-label":"Main"},nt={key:0},ot={key:1,role:"separator",class:"navigation-separator"},at={key:2},st={class:"app-main-content"},it={class:"app-notifications"},rt=["innerHTML"],lt=w({__name:"ApplicationShell",setup(d){const i=z(),u=T(),p=U(),{t:r}=D();return(l,t)=>{const g=m("XTeleportSlot"),v=m("XAction"),k=m("XAlert"),A=m("DataSource"),y=m("XPop"),E=m("XIcon"),N=m("XActionGroup");return c(),S("div",H,[s(g,{name:"modal-layer"}),t[24]||(t[24]=e()),o("header",Y,[o("div",q,[_(l.$slots,"header",{},()=>[s(v,{to:{name:"home"}},{default:n(()=>[_(l.$slots,"home",{},void 0,!0)]),_:3}),t[3]||(t[3]=e()),s(a(P),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:n(()=>t[0]||(t[0]=[e(`
            Star
          `)])),_:1}),t[4]||(t[4]=e()),o("div",Z,[s(A,{src:"/control-plane/version/latest"},{default:n(({data:R})=>[R&&a(u)("KUMA_VERSION")!==R.version?(c(),b(k,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:n(()=>[o("div",j,[o("p",null,f(a(r)("common.product.name"))+` update available
                  `,1),t[2]||(t[2]=e()),s(v,{appearance:"primary",href:a(r)("common.product.href.install")},{default:n(()=>t[1]||(t[1]=[e(`
                    Update
                  `)])),_:1},8,["href"])])]),_:1})):h("",!0)]),_:1})])],!0)]),t[18]||(t[18]=e()),o("div",F,[_(l.$slots,"content-info",{},()=>[o("div",J,[s(y,{width:"280"},{content:n(()=>[o("p",null,[e(f(a(r)("common.product.name"))+" ",1),o("b",null,f(a(u)("KUMA_VERSION")),1),t[6]||(t[6]=e(" on ")),o("b",null,f(a(r)(`common.product.environment.${a(u)("KUMA_ENVIRONMENT")}`)),1),e(" ("+f(a(r)(`common.product.mode.${a(u)("KUMA_MODE")}`))+`)
                `,1)])]),default:n(()=>[s(v,{appearance:"tertiary"},{default:n(()=>t[5]||(t[5]=[e(`
                Info
              `)])),_:1}),t[7]||(t[7]=e())]),_:1})]),t[16]||(t[16]=e()),o("p",Q,[e(f(a(r)("common.product.name"))+" ",1),o("b",null,f(a(u)("KUMA_VERSION")),1),t[8]||(t[8]=e(" on ")),o("b",null,f(a(r)(`common.product.environment.${a(u)("KUMA_ENVIRONMENT")}`)),1),e(" ("+f(a(r)(`common.product.mode.${a(u)("KUMA_MODE")}`))+`)
          `,1)]),t[17]||(t[17]=e()),s(N,null,{control:n(()=>[s(v,{appearance:"tertiary"},{default:n(()=>[s(E,{name:"help"},{default:n(()=>t[9]||(t[9]=[e(`
                  Help
                `)])),_:1})]),_:1})]),default:n(()=>[t[13]||(t[13]=e()),s(v,{href:a(r)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[10]||(t[10]=[e(`
              Documentation
            `)])),_:1},8,["href"]),t[14]||(t[14]=e()),s(v,{href:a(r)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[11]||(t[11]=[e(`
              Feedback
            `)])),_:1},8,["href"]),t[15]||(t[15]=e()),s(v,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[12]||(t[12]=[e(`
              Onboarding
            `)])),_:1})]),_:1})],!0)])]),t[25]||(t[25]=e()),o("div",W,[o("div",tt,[o("nav",et,[i.navigation?(c(),S("ul",nt,[_(l.$slots,"navigation",{},void 0,!0)])):h("",!0),t[19]||(t[19]=e()),i.navigation&&i.bottomNavigation?(c(),S("div",ot)):h("",!0),t[20]||(t[20]=e()),i.bottomNavigation?(c(),S("ul",at,[_(l.$slots,"bottomNavigation",{},void 0,!0)])):h("",!0)])]),t[23]||(t[23]=e()),o("main",st,[o("div",it,[_(l.$slots,"notifications",{},void 0,!0),t[21]||(t[21]=e()),a(p)("use state")?h("",!0):(c(),b(k,{key:0,class:"mb-4",appearance:"warning"},{default:n(()=>[o("ul",null,[o("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:a(r)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,rt)])]),_:1}))]),t[22]||(t[22]=e()),_(l.$slots,"default",{},void 0,!0)])])])}}}),dt=M(lt,[["__scopeId","data-v-60eb680b"]]),ut=["alt"],pt=w({__name:"App",setup(d){var r;const i=L(),u=((r=i.getRoutes().find(l=>l.name==="app"))==null?void 0:r.children.map(l=>(l.name=String(l.name),l)))??[],p=B({name:""});return i.afterEach(()=>{const l=i.currentRoute.value.matched.map(g=>g.name),t=u.find(g=>l.includes(g.name));t&&t.name!==p.value.name&&(p.value=t)}),(l,t)=>{const g=m("RouterView"),v=m("AppView"),k=m("RouteView"),A=m("DataSource");return c(),b(A,{src:"/control-plane/addresses"},{default:n(({data:y})=>[typeof y<"u"?(c(),b(k,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:n(({t:E,can:N})=>[s(dt,{class:"kuma-application"},{home:n(()=>[o("img",{class:"logo",src:x,alt:`${E("common.product.name")} Logo`,"data-testid":"logo"},null,8,ut)]),navigation:n(()=>[s($,{"data-testid":"control-planes-navigator",active:p.value.name==="home",label:"Home",to:{name:"home"},style:{"--icon":"var(--icon-home)"}},null,8,["active"]),t[0]||(t[0]=e()),N("use zones")?(c(),b($,{key:0,"data-testid":"zones-navigator",active:p.value.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"},style:{"--icon":"var(--icon-zones)"}},null,8,["active"])):(c(),b($,{key:1,"data-testid":"zone-egresses-navigator",active:p.value.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"},style:{"--icon":"var(--icon-zone-egresses)"}},null,8,["active"])),t[1]||(t[1]=e()),s($,{active:p.value.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"},style:{"--icon":"var(--icon-meshes)"}},null,8,["active"])]),bottomNavigation:n(()=>[s($,{active:p.value.name==="configuration-view","data-testid":"configuration-navigator",label:"Configuration",to:{name:"configuration-view"},style:{"--icon":"var(--icon-configuration)"}},null,8,["active"])]),default:n(()=>[t[2]||(t[2]=e()),t[3]||(t[3]=e()),t[4]||(t[4]=e()),s(v,null,{default:n(()=>[s(g)]),_:1})]),_:2},1024)]),_:1})):h("",!0)]),_:1})}}}),ct=M(pt,[["__scopeId","data-v-995a408f"]]);export{ct as default};
