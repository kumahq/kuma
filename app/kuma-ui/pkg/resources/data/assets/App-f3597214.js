import{_ as S,o as l,c as A,r as m,d as h,a as _,b as d,w as t,e as c,f as e,n as w,h as C,g as R,i as O,j as z,u as K,k as M,l as a,m as s,t as u,p as b,q as V,s as U,H as D,v as B,x as T}from"./index-a963f507.js";import{K as E}from"./index-fce48c05.js";const H=""+new URL("product-logo-7a2ca341.png",import.meta.url).href,P={},Z={class:"app-navigator"};function x(r,o){return l(),A("li",Z,[m(r.$slots,"default")])}const y=S(P,[["render",x]]),G=h({__name:"ControlPlaneNavigator",setup(r){return(o,p)=>{const i=_("RouterLink");return l(),d(y,{"data-testid":"control-planes-navigator"},{default:t(()=>[c(i,{class:w({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="home")}),to:{name:"home"}},{default:t(()=>[e(`
      Home
    `)]),_:1},8,["class"])]),_:1})}}}),Y=h({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const r={ref:"_"};for(const o in this.$props)r[C(o)]=this.$props[o];return R("span",[O(this.$slots,"default")?R("a",r,this.$slots.default()):R("a",r)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){const r=this.$el.appendChild(document.createElement("span")),o=this;z(()=>import("./buttons.esm-48f94bc9.js"),[],import.meta.url).then(function(p){p.render(r.appendChild(o.$refs._),function(i){try{r.parentNode.replaceChild(i,r)}catch{}})})},reset:function(){this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),q={class:"alert-content"},j=h({__name:"UpgradeCheck",setup(r){const o=K(),{t:p}=M();return(i,n)=>{const k=_("KButton"),g=_("KAlert"),f=_("DataSource");return l(),d(f,{src:"/control-plane/version/latest"},{default:t(({data:v})=>[a(o)("KUMA_VERSION")!==(v==null?void 0:v.version)?(l(),d(g,{key:0,"data-testid":"upgrade-check",class:"upgrade-check-alert",appearance:"info",size:"small"},{alertMessage:t(()=>[s("div",q,[s("div",null,u(a(p)("common.product.name"))+` update available
          `,1),e(),s("div",null,[c(k,{appearance:"primary",to:a(p)("common.product.href.install")},{default:t(()=>[e(`
              Update
            `)]),_:1},8,["to"])])])]),_:1})):b("",!0)]),_:1})}}});const F=S(j,[["__scopeId","data-v-989f5997"]]),I=r=>(B("data-v-b0fef192"),r=r(),T(),r),J={class:"application-shell"},Q={role:"banner"},W={class:"horizontal-list"},X={class:"upgrade-check-wrapper"},ee={class:"horizontal-list"},te={class:"app-status app-status--mobile"},ne={class:"app-status app-status--desktop"},oe=I(()=>s("span",{class:"visually-hidden"},"Help",-1)),ae=I(()=>s("span",{class:"visually-hidden"},"Diagnostics",-1)),se={class:"app-content-container"},ce={key:0,"aria-label":"Main",class:"app-sidebar"},ie={class:"app-main-content"},re={class:"app-notifications"},_e=["innerHTML"],le=h({__name:"ApplicationShell",setup(r){const o=K(),p=V(),{t:i}=M();return(n,k)=>{const g=_("RouterLink"),f=_("KButton"),v=_("KPop"),$=_("KDropdownItem"),L=_("KDropdown"),N=_("KAlert");return l(),A("div",J,[s("header",Q,[s("div",W,[m(n.$slots,"header",{},()=>[c(g,{to:{name:"home"}},{default:t(()=>[m(n.$slots,"home",{},void 0,!0)]),_:3}),e(),c(a(Y),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:t(()=>[e(`
            Star
          `)]),_:1}),e(),s("div",X,[c(F)])],!0)]),e(),s("div",ee,[m(n.$slots,"content-info",{},()=>[s("div",te,[c(v,{width:"280"},{content:t(()=>[s("p",null,[e(u(a(i)("common.product.name"))+" ",1),s("b",null,u(a(o)("KUMA_VERSION")),1),e(" on "),s("b",null,u(a(i)(`common.product.environment.${a(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(a(i)(`common.product.mode.${a(o)("KUMA_MODE")}`))+`)
                `,1)])]),default:t(()=>[c(f,{appearance:"tertiary"},{default:t(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),s("p",ne,[e(u(a(i)("common.product.name"))+" ",1),s("b",null,u(a(o)("KUMA_VERSION")),1),e(" on "),s("b",null,u(a(i)(`common.product.environment.${a(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(a(i)(`common.product.mode.${a(o)("KUMA_MODE")}`))+`)
          `,1)]),e(),c(L,{"kpop-attributes":{placement:"bottomEnd"}},{items:t(()=>[c($,{item:{to:a(i)("common.product.href.docs.index"),label:""},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
                Documentation
              `)]),_:1},8,["item"]),e(),c($,{item:{to:a(i)("common.product.href.feedback"),label:""},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
                Feedback
              `)]),_:1},8,["item"]),e(),c($,{item:{to:{name:"onboarding-welcome-view"},label:""}},{default:t(()=>[e(`
                Onboarding
              `)]),_:1})]),default:t(()=>[c(f,{appearance:"tertiary","icon-only":""},{default:t(()=>[c(a(U),{size:a(E)},null,8,["size"]),e(),oe]),_:1}),e()]),_:1}),e(),c(f,{to:{name:"diagnostics"},appearance:"tertiary","icon-only":"","data-testid":"nav-item-diagnostics"},{default:t(()=>[c(a(D),{size:a(E),"hide-title":""},null,8,["size"]),e(),ae]),_:1})],!0)])]),e(),s("div",se,[n.$slots.navigation?(l(),A("nav",ce,[s("ul",null,[m(n.$slots,"navigation",{},void 0,!0)])])):b("",!0),e(),s("div",ie,[s("div",re,[m(n.$slots,"notifications",{},void 0,!0)]),e(),m(n.$slots,"notifications",{},()=>[a(p)("use state")?b("",!0):(l(),d(N,{key:0,class:"mb-4",appearance:"warning"},{alertMessage:t(()=>[s("ul",null,[s("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:a(i)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,_e),e()])]),_:1}))],!0),e(),m(n.$slots,"default",{},void 0,!0)])])])}}});const pe=S(le,[["__scopeId","data-v-b0fef192"]]),de=h({__name:"MeshNavigator",setup(r){return(o,p)=>{const i=_("RouterLink");return l(),d(y,{"data-testid":"meshes-navigator"},{default:t(()=>[c(i,{class:w({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="mesh-index-view")}),to:{name:"mesh-list-view"}},{default:t(()=>[e(`
      Meshes
    `)]),_:1},8,["class"])]),_:1})}}}),ue=h({__name:"ZoneEgressNavigator",setup(r){return(o,p)=>{const i=_("RouterLink");return l(),d(y,{"data-testid":"zone-egresses-navigator"},{default:t(()=>[c(i,{class:w({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="zone-egress-list-view")}),to:{name:"zone-egress-list-view"}},{default:t(()=>[e(`
      Zone Egresses
    `)]),_:1},8,["class"])]),_:1})}}}),me=h({__name:"ZoneNavigator",setup(r){return(o,p)=>{const i=_("RouterLink");return l(),d(y,{"data-testid":"zones-navigator"},{default:t(()=>[c(i,{class:w({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="zone-index-view")}),to:{name:"zone-cp-list-view"}},{default:t(()=>[e(`
      Zones
    `)]),_:1},8,["class"])]),_:1})}}}),he=["alt"],fe=h({__name:"App",setup(r){return(o,p)=>{const i=_("RouterView"),n=_("AppView"),k=_("RouteView"),g=_("DataSource");return l(),d(g,{src:"/control-plane/addresses"},{default:t(({data:f})=>[typeof f<"u"?(l(),d(k,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:t(({t:v,can:$})=>[c(pe,{class:"kuma-application"},{home:t(()=>[s("img",{class:"logo",src:H,alt:`${v("common.product.name")} Logo`,"data-testid":"logo"},null,8,he)]),navigation:t(()=>[c(G),e(),$("use zones")?(l(),d(me,{key:0})):(l(),d(ue,{key:1})),e(),c(de)]),default:t(()=>[e(),e(),c(n,null,{default:t(()=>[c(i)]),_:1})]),_:2},1024)]),_:1})):b("",!0)]),_:1})}}});const $e=S(fe,[["__scopeId","data-v-f821200e"]]);export{$e as default};
