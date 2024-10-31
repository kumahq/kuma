import{d as A,o as _,c as R,r as m,a as o,w as t,b as e,t as p,n as O,e as d,h as V,f as C,g as X,_ as N,u as I,i as T,j as z,k as r,l as s,m as v,p as k,q as y,s as D,v as U}from"./index-DS1MbHFW.js";const L=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,B={class:"app-navigator"},S=A({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(l){const n=l;return(h,a)=>{const c=d("XAction");return _(),R("li",B,[m(h.$slots,"default",{},()=>[o(c,{class:O({"is-active":n.active}),to:n.to},{default:t(()=>[e(p(n.label),1)]),_:1},8,["class","to"])])])}}}),K=A({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const l={ref:"_"};for(const n in this.$props)l[V(n)]=this.$props[n];return C("span",[X(this.$slots,"default")?C("a",l,this.$slots.default()):C("a",l)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){if(this.$el.lastChild!==this.$refs._)return;const l=this.$el.appendChild(document.createElement("span")),n=this;N(()=>import("./buttons.esm-Bog6bH3O.js"),[],import.meta.url).then(function(h){n.$el.lastChild===l&&h.render(l.appendChild(n.$refs._),function(a){n.$el.lastChild===l&&l.parentNode.replaceChild(a,l)})})},reset:function(){this.$refs._!=null&&this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),G={class:"application-shell"},P={role:"banner"},x={class:"horizontal-list"},H={class:"upgrade-check-wrapper"},Y={class:"alert-content"},q={class:"horizontal-list"},Z={class:"app-status app-status--mobile"},j={class:"app-status app-status--desktop"},F={class:"app-content-container"},J={key:0,"aria-label":"Main",class:"app-sidebar"},Q={class:"app-main-content"},W={class:"app-notifications"},ee=["innerHTML"],te=A({__name:"ApplicationShell",setup(l){const n=I(),h=T(),{t:a}=z();return(c,u)=>{const f=d("XTeleportSlot"),i=d("XAction"),g=d("XAlert"),w=d("DataSource"),E=d("XPop"),b=d("XIcon"),M=d("XActionGroup");return _(),R("div",G,[o(f,{name:"modal-layer"}),e(),r("header",P,[r("div",x,[m(c.$slots,"header",{},()=>[o(i,{to:{name:"home"}},{default:t(()=>[m(c.$slots,"home",{},void 0,!0)]),_:3}),e(),o(s(K),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:t(()=>[e(`
            Star
          `)]),_:1}),e(),r("div",H,[o(w,{src:"/control-plane/version/latest"},{default:t(({data:$})=>[$&&s(n)("KUMA_VERSION")!==$.version?(_(),v(g,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:t(()=>[r("div",Y,[r("p",null,p(s(a)("common.product.name"))+` update available
                  `,1),e(),o(i,{appearance:"primary",href:s(a)("common.product.href.install")},{default:t(()=>[e(`
                    Update
                  `)]),_:1},8,["href"])])]),_:1})):k("",!0)]),_:1})])],!0)]),e(),r("div",q,[m(c.$slots,"content-info",{},()=>[r("div",Z,[o(E,{width:"280"},{content:t(()=>[r("p",null,[e(p(s(a)("common.product.name"))+" ",1),r("b",null,p(s(n)("KUMA_VERSION")),1),e(" on "),r("b",null,p(s(a)(`common.product.environment.${s(n)("KUMA_ENVIRONMENT")}`)),1),e(" ("+p(s(a)(`common.product.mode.${s(n)("KUMA_MODE")}`))+`)
                `,1)])]),default:t(()=>[o(i,{appearance:"tertiary"},{default:t(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),r("p",j,[e(p(s(a)("common.product.name"))+" ",1),r("b",null,p(s(n)("KUMA_VERSION")),1),e(" on "),r("b",null,p(s(a)(`common.product.environment.${s(n)("KUMA_ENVIRONMENT")}`)),1),e(" ("+p(s(a)(`common.product.mode.${s(n)("KUMA_MODE")}`))+`)
          `,1)]),e(),o(M,null,{control:t(()=>[o(i,{appearance:"tertiary"},{default:t(()=>[o(b,{name:"help"},{default:t(()=>[e(`
                  Help
                `)]),_:1})]),_:1})]),default:t(()=>[e(),o(i,{href:s(a)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Documentation
            `)]),_:1},8,["href"]),e(),o(i,{href:s(a)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Feedback
            `)]),_:1},8,["href"]),e(),o(i,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Onboarding
            `)]),_:1})]),_:1}),e(),o(i,{to:{name:"diagnostics"},appearance:"tertiary",icon:"","data-testid":"nav-item-diagnostics"},{default:t(()=>[o(b,{name:"settings"},{default:t(()=>[e(`
              Diagnostics
            `)]),_:1})]),_:1})],!0)])]),e(),r("div",F,[c.$slots.navigation?(_(),R("nav",J,[r("ul",null,[m(c.$slots,"navigation",{},void 0,!0)])])):k("",!0),e(),r("main",Q,[r("div",W,[m(c.$slots,"notifications",{},void 0,!0)]),e(),m(c.$slots,"notifications",{},()=>[s(h)("use state")?k("",!0):(_(),v(g,{key:0,class:"mb-4",appearance:"warning"},{default:t(()=>[r("ul",null,[r("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:s(a)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,ee)])]),_:1}))],!0),e(),m(c.$slots,"default",{},void 0,!0)])])])}}}),ne=y(te,[["__scopeId","data-v-b1282988"]]),ae=["alt"],oe=A({__name:"App",setup(l){var c;const n=D(),h=((c=n.getRoutes().find(u=>u.name==="app"))==null?void 0:c.children.map(u=>(u.name=String(u.name),u)))??[],a=U({name:""});return n.afterEach(()=>{const u=n.currentRoute.value.matched.map(i=>i.name),f=h.find(i=>u.includes(i.name));f&&f.name!==a.value.name&&(a.value=f)}),(u,f)=>{const i=d("RouterView"),g=d("AppView"),w=d("RouteView"),E=d("DataSource");return _(),v(E,{src:"/control-plane/addresses"},{default:t(({data:b})=>[typeof b<"u"?(_(),v(w,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:t(({t:M,can:$})=>[o(ne,{class:"kuma-application"},{home:t(()=>[r("img",{class:"logo",src:L,alt:`${M("common.product.name")} Logo`,"data-testid":"logo"},null,8,ae)]),navigation:t(()=>[o(S,{"data-testid":"control-planes-navigator",active:a.value.name==="home",label:"Home",to:{name:"home"}},null,8,["active"]),e(),$("use zones")?(_(),v(S,{key:0,"data-testid":"zones-navigator",active:a.value.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"}},null,8,["active"])):(_(),v(S,{key:1,"data-testid":"zone-egresses-navigator",active:a.value.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"}},null,8,["active"])),e(),o(S,{active:a.value.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"}},null,8,["active"])]),default:t(()=>[e(),e(),o(g,null,{default:t(()=>[o(i)]),_:1})]),_:2},1024)]),_:1})):k("",!0)]),_:1})}}}),re=y(oe,[["__scopeId","data-v-5bc263b6"]]);export{re as default};
