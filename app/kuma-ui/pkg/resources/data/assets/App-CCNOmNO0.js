import{d as $,r as i,o as p,c as v,a as u,b as n,w as t,e,t as d,n as I,h as K,f as y,g as L,_ as T,u as X,i as z,j as D,k as a,l as o,m as h,p as S,q as C,s as U,F as V}from"./index-DOjcqG3h.js";const B=""+new URL("product-logo-CDoXkXpC.png",import.meta.url).href,G={class:"app-navigator"},k=$({__name:"AppNavigator",props:{active:{type:Boolean,default:!1},label:{default:""},to:{default:()=>({})}},setup(l){const s=l;return(m,r)=>{const c=i("XAction");return p(),v("li",G,[u(m.$slots,"default",{},()=>[n(c,{class:I({"is-active":s.active}),to:s.to},{default:t(()=>[e(d(s.label),1)]),_:1},8,["class","to"])])])}}}),P=$({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const l={ref:"_"};for(const s in this.$props)l[K(s)]=this.$props[s];return y("span",[L(this.$slots,"default")?y("a",l,this.$slots.default()):y("a",l)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){const l=this.$el.appendChild(document.createElement("span")),s=this;T(()=>import("./buttons.esm-DQonl2ce.js"),[],import.meta.url).then(function(m){m.render(l.appendChild(s.$refs._),function(r){try{l.parentNode.replaceChild(r,l)}catch{}})})},reset:function(){this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),x={class:"application-shell"},H={role:"banner"},Y={class:"horizontal-list"},q={class:"upgrade-check-wrapper"},F={class:"alert-content"},Z={class:"horizontal-list"},j={class:"app-status app-status--mobile"},J={class:"app-status app-status--desktop"},Q={class:"app-content-container"},W={key:0,"aria-label":"Main",class:"app-sidebar"},ee={class:"app-main-content"},te={class:"app-notifications"},ne=["innerHTML"],oe=$({__name:"ApplicationShell",setup(l){const s=X(),m=z(),{t:r}=D();return(c,M)=>{const A=i("XTeleportSlot"),w=i("RouterLink"),f=i("KButton"),g=i("KAlert"),E=i("DataSource"),_=i("KPop"),R=i("XIcon"),b=i("XAction"),N=i("XActionGroup");return p(),v("div",x,[n(A,{name:"modal-layer"}),e(),a("header",H,[a("div",Y,[u(c.$slots,"header",{},()=>[n(w,{to:{name:"home"}},{default:t(()=>[u(c.$slots,"home",{},void 0,!0)]),_:3}),e(),n(o(P),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:t(()=>[e(`
            Star
          `)]),_:1}),e(),a("div",q,[n(E,{src:"/control-plane/version/latest"},{default:t(({data:O})=>[O&&o(s)("KUMA_VERSION")!==O.version?(p(),h(g,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:t(()=>[a("div",F,[a("p",null,d(o(r)("common.product.name"))+` update available
                  `,1),e(),n(f,{appearance:"primary",to:o(r)("common.product.href.install")},{default:t(()=>[e(`
                    Update
                  `)]),_:1},8,["to"])])]),_:1})):S("",!0)]),_:1})])],!0)]),e(),a("div",Z,[u(c.$slots,"content-info",{},()=>[a("div",j,[n(_,{width:"280"},{content:t(()=>[a("p",null,[e(d(o(r)("common.product.name"))+" ",1),a("b",null,d(o(s)("KUMA_VERSION")),1),e(" on "),a("b",null,d(o(r)(`common.product.environment.${o(s)("KUMA_ENVIRONMENT")}`)),1),e(" ("+d(o(r)(`common.product.mode.${o(s)("KUMA_MODE")}`))+`)
                `,1)])]),default:t(()=>[n(f,{appearance:"tertiary"},{default:t(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),a("p",J,[e(d(o(r)("common.product.name"))+" ",1),a("b",null,d(o(s)("KUMA_VERSION")),1),e(" on "),a("b",null,d(o(r)(`common.product.environment.${o(s)("KUMA_ENVIRONMENT")}`)),1),e(" ("+d(o(r)(`common.product.mode.${o(s)("KUMA_MODE")}`))+`)
          `,1)]),e(),n(N,null,{control:t(()=>[n(b,{appearance:"tertiary"},{default:t(()=>[n(R,{name:"help"},{default:t(()=>[e(`
                  Help
                `)]),_:1})]),_:1})]),default:t(()=>[e(),n(b,{href:o(r)("common.product.href.docs.index"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Documentation
            `)]),_:1},8,["href"]),e(),n(b,{href:o(r)("common.product.href.feedback"),target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Feedback
            `)]),_:1},8,["href"]),e(),n(b,{to:{name:"onboarding-welcome-view"},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
              Onboarding
            `)]),_:1})]),_:1}),e(),n(f,{to:{name:"diagnostics"},appearance:"tertiary",icon:"","data-testid":"nav-item-diagnostics"},{default:t(()=>[n(R,{name:"settings"},{default:t(()=>[e(`
              Diagnostics
            `)]),_:1})]),_:1})],!0)])]),e(),a("div",Q,[c.$slots.navigation?(p(),v("nav",W,[a("ul",null,[u(c.$slots,"navigation",{},void 0,!0)])])):S("",!0),e(),a("main",ee,[a("div",te,[u(c.$slots,"notifications",{},void 0,!0)]),e(),u(c.$slots,"notifications",{},()=>[o(m)("use state")?S("",!0):(p(),h(g,{key:0,class:"mb-4",appearance:"warning"},{default:t(()=>[a("ul",null,[a("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:o(r)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,ne)])]),_:1}))],!0),e(),u(c.$slots,"default",{},void 0,!0)])])])}}}),ae=C(oe,[["__scopeId","data-v-228556cd"]]),se=["alt"],re=$({__name:"App",setup(l){return(s,m)=>{const r=i("RouterView"),c=i("AppView"),M=i("RouteView"),A=i("DataSource");return p(),h(A,{src:"/control-plane/addresses"},{default:t(({data:w})=>[typeof w<"u"?(p(),h(M,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:t(({t:f,can:g,route:E})=>[n(ae,{class:"kuma-application"},{home:t(()=>[a("img",{class:"logo",src:B,alt:`${f("common.product.name")} Logo`,"data-testid":"logo"},null,8,se)]),navigation:t(()=>[(p(!0),v(V,null,U([E.child()??{name:""}],_=>(p(),v(V,{key:_.name},[n(k,{"data-testid":"control-planes-navigator",active:_.name==="home",label:"Home",to:{name:"home"}},null,8,["active"]),e(),g("use zones")?(p(),h(k,{key:0,"data-testid":"zones-navigator",active:_.name==="zone-index-view",label:"Zones",to:{name:"zone-index-view"}},null,8,["active"])):(p(),h(k,{key:1,"data-testid":"zone-egresses-navigator",active:_.name==="zone-egress-index-view",label:"Zone Egresses",to:{name:"zone-egress-list-view"}},null,8,["active"])),e(),n(k,{active:_.name==="mesh-index-view","data-testid":"meshes-navigator",label:"Meshes",to:{name:"mesh-index-view"}},null,8,["active"])],64))),128))]),default:t(()=>[e(),e(),n(c,null,{default:t(()=>[n(r)]),_:1})]),_:2},1024)]),_:1})):S("",!0)]),_:1})}}}),ce=C(re,[["__scopeId","data-v-5f44d1a6"]]);export{ce as default};
