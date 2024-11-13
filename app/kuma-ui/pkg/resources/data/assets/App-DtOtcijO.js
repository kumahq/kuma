import{d as k,o as m,c as C,r as h,a as o,w as n,b as e,t as u,n as R,e as p,h as O,_ as V,u as X,f as N,g as I,i as r,j as s,k as v,l as w,m as M,p as T,q as z}from"./index-loxRIpgb.js";const D=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,L={class:"app-navigator"},S=k({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(i){const t=i;return(d,a)=>{const l=p("XAction");return m(),C("li",L,[h(d.$slots,"default",{},()=>[o(l,{class:R({"is-active":t.active}),to:t.to},{default:n(()=>[e(u(t.label),1)]),_:1},8,["class","to"])])])}}});/**
* @vue/shared v3.5.12
* (c) 2018-present Yuxi (Evan) You and Vue contributors
* @license MIT
**/const U=Object.prototype.hasOwnProperty,B=(i,t)=>U.call(i,t),P=i=>{const t=Object.create(null);return d=>t[d]||(t[d]=i(d))},K=/\B([A-Z])/g,G=P(i=>i.replace(K,"-$1").toLowerCase()),x=k({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const i={ref:"_"};for(const t in this.$props)i[G(t)]=this.$props[t];return O("span",[B(this.$slots,"default")?O("a",i,this.$slots.default()):O("a",i)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){if(this.$el.lastChild!==this.$refs._)return;const i=this.$el.appendChild(document.createElement("span")),t=this;V(()=>import("./buttons.esm-DK2fWHEW.js"),[],import.meta.url).then(function(d){t.$el.lastChild===i&&d.render(i.appendChild(t.$refs._),function(a){t.$el.lastChild===i&&i.parentNode.replaceChild(a,i)})})},reset:function(){this.$refs._!=null&&this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),H={class:"application-shell"},Y={role:"banner"},j={class:"horizontal-list"},q={class:"upgrade-check-wrapper"},Z={class:"alert-content"},F={class:"horizontal-list"},J={class:"app-status app-status--mobile"},Q={class:"app-status app-status--desktop"},W={class:"app-content-container"},ee={key:0,"aria-label":"Main",class:"app-sidebar"},te={class:"app-main-content"},ne={class:"app-notifications"},ae=["innerHTML"],oe=k({__name:"ApplicationShell",setup(i){const t=X(),d=N(),{t:a}=I();return(l,_)=>{const f=p("XTeleportSlot"),c=p("XAction"),g=p("XAlert"),A=p("DataSource"),E=p("XPop"),b=p("XIcon"),y=p("XActionGroup");return m(),C("div",H,[o(f,{name:"modal-layer"}),e(),r("header",Y,[r("div",j,[h(l.$slots,"header",{},()=>[o(c,{to:{name:"home"}},{default:n(()=>[h(l.$slots,"home",{},void 0,!0)]),_:3}),e(),o(s(x),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:n(()=>[e(`
            Star
          `)]),_:1}),e(),r("div",q,[o(A,{src:"/control-plane/version/latest"},{default:n(({data:$})=>[$&&s(t)("KUMA_VERSION")!==$.version?(m(),v(g,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:n(()=>[r("div",Z,[r("p",null,u(s(a)("common.product.name"))+` update available
                  `,1),e(),o(c,{appearance:"primary",href:s(a)("common.product.href.install")},{default:n(()=>[e(`
                    Update
                  `)]),_:1},8,["href"])])]),_:1})):w("",!0)]),_:1})])],!0)]),e(),r("div",F,[h(l.$slots,"content-info",{},()=>[r("div",J,[o(E,{width:"280"},{content:n(()=>[r("p",null,[e(u(s(a)("common.product.name"))+" ",1),r("b",null,u(s(t)("KUMA_VERSION")),1),e(" on "),r("b",null,u(s(a)(`common.product.environment.${s(t)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(s(a)(`common.product.mode.${s(t)("KUMA_MODE")}`))+`)
                `,1)])]),default:n(()=>[o(c,{appearance:"tertiary"},{default:n(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),r("p",Q,[e(u(s(a)("common.product.name"))+" ",1),r("b",null,u(s(t)("KUMA_VERSION")),1),e(" on "),r("b",null,u(s(a)(`common.product.environment.${s(t)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(s(a)(`common.product.mode.${s(t)("KUMA_MODE")}`))+`)
          `,1)]),e(),o(y,null,{control:n(()=>[o(c,{appearance:"tertiary"},{default:n(()=>[o(b,{name:"help"},{default:n(()=>[e(`
                  Help
                `)]),_:1})]),_:1})]),default:n(()=>[e(),o(c,{href:s(a)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>[e(`
              Documentation
            `)]),_:1},8,["href"]),e(),o(c,{href:s(a)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:n(()=>[e(`
              Feedback
            `)]),_:1},8,["href"]),e(),o(c,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:n(()=>[e(`
              Onboarding
            `)]),_:1})]),_:1}),e(),o(c,{to:{name:"diagnostics"},appearance:"tertiary",icon:"","data-testid":"nav-item-diagnostics"},{default:n(()=>[o(b,{name:"settings"},{default:n(()=>[e(`
              Diagnostics
            `)]),_:1})]),_:1})],!0)])]),e(),r("div",W,[l.$slots.navigation?(m(),C("nav",ee,[r("ul",null,[h(l.$slots,"navigation",{},void 0,!0)])])):w("",!0),e(),r("main",te,[r("div",ne,[h(l.$slots,"notifications",{},void 0,!0)]),e(),h(l.$slots,"notifications",{},()=>[s(d)("use state")?w("",!0):(m(),v(g,{key:0,class:"mb-4",appearance:"warning"},{default:n(()=>[r("ul",null,[r("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:s(a)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,ae)])]),_:1}))],!0),e(),h(l.$slots,"default",{},void 0,!0)])])])}}}),se=M(oe,[["__scopeId","data-v-b1282988"]]),re=["alt"],ie=k({__name:"App",setup(i){var l;const t=T(),d=((l=t.getRoutes().find(_=>_.name==="app"))==null?void 0:l.children.map(_=>(_.name=String(_.name),_)))??[],a=z({name:""});return t.afterEach(()=>{const _=t.currentRoute.value.matched.map(c=>c.name),f=d.find(c=>_.includes(c.name));f&&f.name!==a.value.name&&(a.value=f)}),(_,f)=>{const c=p("RouterView"),g=p("AppView"),A=p("RouteView"),E=p("DataSource");return m(),v(E,{src:"/control-plane/addresses"},{default:n(({data:b})=>[typeof b<"u"?(m(),v(A,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:n(({t:y,can:$})=>[o(se,{class:"kuma-application"},{home:n(()=>[r("img",{class:"logo",src:D,alt:`${y("common.product.name")} Logo`,"data-testid":"logo"},null,8,re)]),navigation:n(()=>[o(S,{"data-testid":"control-planes-navigator",active:a.value.name==="home",label:"Home",to:{name:"home"}},null,8,["active"]),e(),$("use zones")?(m(),v(S,{key:0,"data-testid":"zones-navigator",active:a.value.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"}},null,8,["active"])):(m(),v(S,{key:1,"data-testid":"zone-egresses-navigator",active:a.value.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"}},null,8,["active"])),e(),o(S,{active:a.value.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"}},null,8,["active"])]),default:n(()=>[e(),e(),o(g,null,{default:n(()=>[o(c)]),_:1})]),_:2},1024)]),_:1})):w("",!0)]),_:1})}}}),le=M(ie,[["__scopeId","data-v-5bc263b6"]]);export{le as default};