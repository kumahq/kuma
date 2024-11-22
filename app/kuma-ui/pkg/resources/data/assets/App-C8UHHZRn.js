import{d as A,o as c,c as R,r as _,a as s,w as n,b as e,t as f,n as O,e as m,h as V,f as C,g as X,_ as N,u as I,i as T,j as z,k as i,l as r,m as g,p as k,q as y,s as D,v as U}from"./index-DpJ_igul.js";const L=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,B={class:"app-navigator"},S=A({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(p){const o=p;return(v,a)=>{const u=m("XAction");return c(),R("li",B,[_(v.$slots,"default",{},()=>[s(u,{class:O({"is-active":o.active}),to:o.to},{default:n(()=>[e(f(o.label),1)]),_:1},8,["class","to"])])])}}}),K=A({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const p={ref:"_"};for(const o in this.$props)p[V(o)]=this.$props[o];return C("span",[X(this.$slots,"default")?C("a",p,this.$slots.default()):C("a",p)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){if(this.$el.lastChild!==this.$refs._)return;const p=this.$el.appendChild(document.createElement("span")),o=this;N(()=>import("./buttons.esm-DK2fWHEW.js"),[],import.meta.url).then(function(v){o.$el.lastChild===p&&v.render(p.appendChild(o.$refs._),function(a){o.$el.lastChild===p&&p.parentNode.replaceChild(a,p)})})},reset:function(){this.$refs._!=null&&this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),G={class:"application-shell"},P={role:"banner"},x={class:"horizontal-list"},H={class:"upgrade-check-wrapper"},Y={class:"alert-content"},q={class:"horizontal-list"},Z={class:"app-status app-status--mobile"},j={class:"app-status app-status--desktop"},F={class:"app-content-container"},J={key:0,"aria-label":"Main",class:"app-sidebar"},Q={class:"app-main-content"},W={class:"app-notifications"},tt=["innerHTML"],et=A({__name:"ApplicationShell",setup(p){const o=I(),v=T(),{t:a}=z();return(u,t)=>{const l=m("XTeleportSlot"),d=m("XAction"),h=m("XAlert"),w=m("DataSource"),E=m("XPop"),b=m("XIcon"),M=m("XActionGroup");return c(),R("div",G,[s(l,{name:"modal-layer"}),t[24]||(t[24]=e()),i("header",P,[i("div",x,[_(u.$slots,"header",{},()=>[s(d,{to:{name:"home"}},{default:n(()=>[_(u.$slots,"home",{},void 0,!0)]),_:3}),t[3]||(t[3]=e()),s(r(K),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:n(()=>t[0]||(t[0]=[e(`
            Star
          `)])),_:1}),t[4]||(t[4]=e()),i("div",H,[s(w,{src:"/control-plane/version/latest"},{default:n(({data:$})=>[$&&r(o)("KUMA_VERSION")!==$.version?(c(),g(h,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:n(()=>[i("div",Y,[i("p",null,f(r(a)("common.product.name"))+` update available
                  `,1),t[2]||(t[2]=e()),s(d,{appearance:"primary",href:r(a)("common.product.href.install")},{default:n(()=>t[1]||(t[1]=[e(`
                    Update
                  `)])),_:1},8,["href"])])]),_:1})):k("",!0)]),_:1})])],!0)]),t[20]||(t[20]=e()),i("div",q,[_(u.$slots,"content-info",{},()=>[i("div",Z,[s(E,{width:"280"},{content:n(()=>[i("p",null,[e(f(r(a)("common.product.name"))+" ",1),i("b",null,f(r(o)("KUMA_VERSION")),1),t[6]||(t[6]=e(" on ")),i("b",null,f(r(a)(`common.product.environment.${r(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+f(r(a)(`common.product.mode.${r(o)("KUMA_MODE")}`))+`)
                `,1)])]),default:n(()=>[s(d,{appearance:"tertiary"},{default:n(()=>t[5]||(t[5]=[e(`
                Info
              `)])),_:1}),t[7]||(t[7]=e())]),_:1})]),t[17]||(t[17]=e()),i("p",j,[e(f(r(a)("common.product.name"))+" ",1),i("b",null,f(r(o)("KUMA_VERSION")),1),t[8]||(t[8]=e(" on ")),i("b",null,f(r(a)(`common.product.environment.${r(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+f(r(a)(`common.product.mode.${r(o)("KUMA_MODE")}`))+`)
          `,1)]),t[18]||(t[18]=e()),s(M,null,{control:n(()=>[s(d,{appearance:"tertiary"},{default:n(()=>[s(b,{name:"help"},{default:n(()=>t[9]||(t[9]=[e(`
                  Help
                `)])),_:1})]),_:1})]),default:n(()=>[t[13]||(t[13]=e()),s(d,{href:r(a)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[10]||(t[10]=[e(`
              Documentation
            `)])),_:1},8,["href"]),t[14]||(t[14]=e()),s(d,{href:r(a)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[11]||(t[11]=[e(`
              Feedback
            `)])),_:1},8,["href"]),t[15]||(t[15]=e()),s(d,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:n(()=>t[12]||(t[12]=[e(`
              Onboarding
            `)])),_:1})]),_:1}),t[19]||(t[19]=e()),s(d,{to:{name:"diagnostics"},appearance:"tertiary",icon:"","data-testid":"nav-item-diagnostics"},{default:n(()=>[s(b,{name:"settings"},{default:n(()=>t[16]||(t[16]=[e(`
              Diagnostics
            `)])),_:1})]),_:1})],!0)])]),t[25]||(t[25]=e()),i("div",F,[u.$slots.navigation?(c(),R("nav",J,[i("ul",null,[_(u.$slots,"navigation",{},void 0,!0)])])):k("",!0),t[23]||(t[23]=e()),i("main",Q,[i("div",W,[_(u.$slots,"notifications",{},void 0,!0)]),t[21]||(t[21]=e()),_(u.$slots,"notifications",{},()=>[r(v)("use state")?k("",!0):(c(),g(h,{key:0,class:"mb-4",appearance:"warning"},{default:n(()=>[i("ul",null,[i("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:r(a)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,tt)])]),_:1}))],!0),t[22]||(t[22]=e()),_(u.$slots,"default",{},void 0,!0)])])])}}}),nt=y(et,[["__scopeId","data-v-b1282988"]]),ot=["alt"],at=A({__name:"App",setup(p){var u;const o=D(),v=((u=o.getRoutes().find(t=>t.name==="app"))==null?void 0:u.children.map(t=>(t.name=String(t.name),t)))??[],a=U({name:""});return o.afterEach(()=>{const t=o.currentRoute.value.matched.map(d=>d.name),l=v.find(d=>t.includes(d.name));l&&l.name!==a.value.name&&(a.value=l)}),(t,l)=>{const d=m("RouterView"),h=m("AppView"),w=m("RouteView"),E=m("DataSource");return c(),g(E,{src:"/control-plane/addresses"},{default:n(({data:b})=>[typeof b<"u"?(c(),g(w,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:n(({t:M,can:$})=>[s(nt,{class:"kuma-application"},{home:n(()=>[i("img",{class:"logo",src:L,alt:`${M("common.product.name")} Logo`,"data-testid":"logo"},null,8,ot)]),navigation:n(()=>[s(S,{"data-testid":"control-planes-navigator",active:a.value.name==="home",label:"Home",to:{name:"home"}},null,8,["active"]),l[0]||(l[0]=e()),$("use zones")?(c(),g(S,{key:0,"data-testid":"zones-navigator",active:a.value.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"}},null,8,["active"])):(c(),g(S,{key:1,"data-testid":"zone-egresses-navigator",active:a.value.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"}},null,8,["active"])),l[1]||(l[1]=e()),s(S,{active:a.value.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"}},null,8,["active"])]),default:n(()=>[l[2]||(l[2]=e()),l[3]||(l[3]=e()),s(h,null,{default:n(()=>[s(d)]),_:1})]),_:2},1024)]),_:1})):k("",!0)]),_:1})}}}),rt=y(at,[["__scopeId","data-v-5bc263b6"]]);export{rt as default};
