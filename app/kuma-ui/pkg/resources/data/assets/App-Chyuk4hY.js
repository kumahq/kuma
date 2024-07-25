import{d as w,r as c,o as u,c as y,a as _,b as a,w as t,e,t as d,n as N,h as I,f as M,g as K,_ as T,u as L,i as X,j as z,k as r,l as s,m as g,p as A,q as V,s as D,v as U}from"./index-Bhm2pUpK.js";const B=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,G={class:"app-navigator"},$=w({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(l){const o=l;return(m,n)=>{const i=c("XAction");return u(),y("li",G,[_(m.$slots,"default",{},()=>[a(i,{class:N({"is-active":o.active}),to:o.to},{default:t(()=>[e(d(o.label),1)]),_:1},8,["class","to"])])])}}}),P=w({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const l={ref:"_"};for(const o in this.$props)l[I(o)]=this.$props[o];return M("span",[K(this.$slots,"default")?M("a",l,this.$slots.default()):M("a",l)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){const l=this.$el.appendChild(document.createElement("span")),o=this;T(()=>import("./buttons.esm-DQonl2ce.js"),[],import.meta.url).then(function(m){m.render(l.appendChild(o.$refs._),function(n){try{l.parentNode.replaceChild(n,l)}catch{}})})},reset:function(){this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),x={class:"application-shell"},H={role:"banner"},Y={class:"horizontal-list"},q={class:"upgrade-check-wrapper"},Z={class:"alert-content"},j={class:"horizontal-list"},F={class:"app-status app-status--mobile"},J={class:"app-status app-status--desktop"},Q={class:"app-content-container"},W={key:0,"aria-label":"Main",class:"app-sidebar"},ee={class:"app-main-content"},te={class:"app-notifications"},ne=["innerHTML"],ae=w({__name:"ApplicationShell",setup(l){const o=L(),m=X(),{t:n}=z();return(i,p)=>{const f=c("XTeleportSlot"),h=c("RouterLink"),b=c("KButton"),S=c("KAlert"),E=c("DataSource"),R=c("KPop"),k=c("XIcon"),v=c("XAction"),C=c("XActionGroup");return u(),y("div",x,[a(f,{name:"modal-layer"}),e(),r("header",H,[r("div",Y,[_(i.$slots,"header",{},()=>[a(h,{to:{name:"home"}},{default:t(()=>[_(i.$slots,"home",{},void 0,!0)]),_:3}),e(),a(s(P),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:t(()=>[e(`
            Star
          `)]),_:1}),e(),r("div",q,[a(E,{src:"/control-plane/version/latest"},{default:t(({data:O})=>[O&&s(o)("KUMA_VERSION")!==O.version?(u(),g(S,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:t(()=>[r("div",Z,[r("p",null,d(s(n)("common.product.name"))+` update available
                  `,1),e(),a(b,{appearance:"primary",to:s(n)("common.product.href.install")},{default:t(()=>[e(`
                    Update
                  `)]),_:1},8,["to"])])]),_:1})):A("",!0)]),_:1})])],!0)]),e(),r("div",j,[_(i.$slots,"content-info",{},()=>[r("div",F,[a(R,{width:"280"},{content:t(()=>[r("p",null,[e(d(s(n)("common.product.name"))+" ",1),r("b",null,d(s(o)("KUMA_VERSION")),1),e(" on "),r("b",null,d(s(n)(`common.product.environment.${s(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+d(s(n)(`common.product.mode.${s(o)("KUMA_MODE")}`))+`)
                `,1)])]),default:t(()=>[a(b,{appearance:"tertiary"},{default:t(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),r("p",J,[e(d(s(n)("common.product.name"))+" ",1),r("b",null,d(s(o)("KUMA_VERSION")),1),e(" on "),r("b",null,d(s(n)(`common.product.environment.${s(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+d(s(n)(`common.product.mode.${s(o)("KUMA_MODE")}`))+`)
          `,1)]),e(),a(C,null,{control:t(()=>[a(v,{appearance:"tertiary"},{default:t(()=>[a(k,{name:"help"},{default:t(()=>[e(`
                  Help
                `)]),_:1})]),_:1})]),default:t(()=>[e(),a(v,{href:s(n)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Documentation
            `)]),_:1},8,["href"]),e(),a(v,{href:s(n)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Feedback
            `)]),_:1},8,["href"]),e(),a(v,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Onboarding
            `)]),_:1})]),_:1}),e(),a(b,{to:{name:"diagnostics"},appearance:"tertiary",icon:"","data-testid":"nav-item-diagnostics"},{default:t(()=>[a(k,{name:"settings"},{default:t(()=>[e(`
              Diagnostics
            `)]),_:1})]),_:1})],!0)])]),e(),r("div",Q,[i.$slots.navigation?(u(),y("nav",W,[r("ul",null,[_(i.$slots,"navigation",{},void 0,!0)])])):A("",!0),e(),r("main",ee,[r("div",te,[_(i.$slots,"notifications",{},void 0,!0)]),e(),_(i.$slots,"notifications",{},()=>[s(m)("use state")?A("",!0):(u(),g(S,{key:0,class:"mb-4",appearance:"warning"},{default:t(()=>[r("ul",null,[r("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:s(n)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,ne)])]),_:1}))],!0),e(),_(i.$slots,"default",{},void 0,!0)])])])}}}),oe=V(ae,[["__scopeId","data-v-8d9ae308"]]),se=["alt"],re=w({__name:"App",setup(l){var i;const o=D(),m=((i=o.getRoutes().find(p=>p.name==="app"))==null?void 0:i.children.map(p=>(p.name=String(p.name),p)))??[],n=U({name:""});return o.afterEach(()=>{const p=o.currentRoute.value.matched.map(h=>h.name),f=m.find(h=>p.includes(h.name));f&&f.name!==n.value.name&&(n.value=f)}),(p,f)=>{const h=c("RouterView"),b=c("AppView"),S=c("RouteView"),E=c("DataSource");return u(),g(E,{src:"/control-plane/addresses"},{default:t(({data:R})=>[typeof R<"u"?(u(),g(S,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:t(({t:k,can:v})=>[a(oe,{class:"kuma-application"},{home:t(()=>[r("img",{class:"logo",src:B,alt:`${k("common.product.name")} Logo`,"data-testid":"logo"},null,8,se)]),navigation:t(()=>[a($,{"data-testid":"control-planes-navigator",active:n.value.name==="home",label:"Home",to:{name:"home"}},null,8,["active"]),e(),v("use zones")?(u(),g($,{key:0,"data-testid":"zones-navigator",active:n.value.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"}},null,8,["active"])):(u(),g($,{key:1,"data-testid":"zone-egresses-navigator",active:n.value.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"}},null,8,["active"])),e(),a($,{active:n.value.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"}},null,8,["active"])]),default:t(()=>[e(),e(),a(b,null,{default:t(()=>[a(h)]),_:1})]),_:2},1024)]),_:1})):A("",!0)]),_:1})}}}),ce=V(re,[["__scopeId","data-v-5bc263b6"]]);export{ce as default};
