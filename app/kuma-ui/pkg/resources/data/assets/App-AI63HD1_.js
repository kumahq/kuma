function __vite__mapDeps(indexes) {
  if (!__vite__mapDeps.viteFileDeps) {
    __vite__mapDeps.viteFileDeps = []
  }
  return indexes.map((i) => __vite__mapDeps.viteFileDeps[i])
}
import{_ as y,o as p,c as E,r as m,d as h,a as l,b as d,w as t,e as s,f as e,n as $,h as I,g as R,i as O,j as V,u as C,k as D,l as z,m as i,p as r,t as u,q as g,s as T,v as U,x as B,y as P}from"./index-CP9JG8i6.js";const q=""+new URL("product-logo-g6F5F6Qq.png",import.meta.url).href,H={},G={class:"app-navigator"};function Y(c,o){return p(),E("li",G,[m(c.$slots,"default")])}const k=y(H,[["render",Y]]),Z=h({__name:"ControlPlaneNavigator",setup(c){return(o,_)=>{const a=l("RouterLink");return p(),d(k,{"data-testid":"control-planes-navigator"},{default:t(()=>[s(a,{class:$({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="home")}),to:{name:"home"}},{default:t(()=>[e(`
      Home
    `)]),_:1},8,["class"])]),_:1})}}}),F=h({name:"github-button",props:{href:String,ariaLabel:String,title:String,dataIcon:String,dataColorScheme:String,dataSize:String,dataShowCount:String,dataText:String},render:function(){const c={ref:"_"};for(const o in this.$props)c[I(o)]=this.$props[o];return R("span",[O(this.$slots,"default")?R("a",c,this.$slots.default()):R("a",c)])},mounted:function(){this.paint()},beforeUpdate:function(){this.reset()},updated:function(){this.paint()},beforeUnmount:function(){this.reset()},methods:{paint:function(){const c=this.$el.appendChild(document.createElement("span")),o=this;V(()=>import("./buttons.esm-TXzUqbYj.js"),__vite__mapDeps([]),import.meta.url).then(function(_){_.render(c.appendChild(o.$refs._),function(a){try{c.parentNode.replaceChild(a,c)}catch{}})})},reset:function(){this.$el.replaceChild(this.$refs._,this.$el.lastChild)}}}),A=c=>(B("data-v-c98e0e53"),c=c(),P(),c),j={class:"application-shell"},Q={role:"banner"},J={class:"horizontal-list"},W={class:"upgrade-check-wrapper"},X={class:"alert-content"},x={class:"horizontal-list"},ee={class:"app-status app-status--mobile"},te={class:"app-status app-status--desktop"},ne=A(()=>i("span",{class:"visually-hidden"},"Help",-1)),oe=A(()=>i("span",{class:"visually-hidden"},"Diagnostics",-1)),ae={class:"app-content-container"},se={key:0,"aria-label":"Main",class:"app-sidebar"},re={class:"app-main-content"},ie={class:"app-notifications"},ce=["innerHTML"],le=h({__name:"ApplicationShell",setup(c){const o=C(),_=D(),{t:a}=z();return(n,L)=>{const b=l("RouterLink"),f=l("KButton"),v=l("KAlert"),w=l("DataSource"),N=l("KPop"),S=l("KDropdownItem"),K=l("KDropdown");return p(),E("div",j,[i("header",Q,[i("div",J,[m(n.$slots,"header",{},()=>[s(b,{to:{name:"home"}},{default:t(()=>[m(n.$slots,"home",{},void 0,!0)]),_:3}),e(),s(r(F),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:t(()=>[e(`
            Star
          `)]),_:1}),e(),i("div",W,[s(w,{src:"/control-plane/version/latest"},{default:t(({data:M})=>[M&&r(o)("KUMA_VERSION")!==M.version?(p(),d(v,{key:0,class:"upgrade-alert","data-testid":"upgrade-check",appearance:"info"},{default:t(()=>[i("div",X,[i("p",null,u(r(a)("common.product.name"))+` update available
                  `,1),e(),s(f,{appearance:"primary",to:r(a)("common.product.href.install")},{default:t(()=>[e(`
                    Update
                  `)]),_:1},8,["to"])])]),_:1})):g("",!0)]),_:1})])],!0)]),e(),i("div",x,[m(n.$slots,"content-info",{},()=>[i("div",ee,[s(N,{width:"280"},{content:t(()=>[i("p",null,[e(u(r(a)("common.product.name"))+" ",1),i("b",null,u(r(o)("KUMA_VERSION")),1),e(" on "),i("b",null,u(r(a)(`common.product.environment.${r(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(r(a)(`common.product.mode.${r(o)("KUMA_MODE")}`))+`)
                `,1)])]),default:t(()=>[s(f,{appearance:"tertiary"},{default:t(()=>[e(`
                Info
              `)]),_:1}),e()]),_:1})]),e(),i("p",te,[e(u(r(a)("common.product.name"))+" ",1),i("b",null,u(r(o)("KUMA_VERSION")),1),e(" on "),i("b",null,u(r(a)(`common.product.environment.${r(o)("KUMA_ENVIRONMENT")}`)),1),e(" ("+u(r(a)(`common.product.mode.${r(o)("KUMA_MODE")}`))+`)
          `,1)]),e(),s(K,{"kpop-attributes":{placement:"bottomEnd"}},{items:t(()=>[s(S,{item:{to:r(a)("common.product.href.docs.index"),label:""},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
                Documentation
              `)]),_:1},8,["item"]),e(),s(S,{item:{to:r(a)("common.product.href.feedback"),label:""},target:"_blank",rel:"noopener noreferrer"},{default:t(()=>[e(`
                Feedback
              `)]),_:1},8,["item"]),e(),s(S,{item:{to:{name:"onboarding-welcome-view"},label:""}},{default:t(()=>[e(`
                Onboarding
              `)]),_:1})]),default:t(()=>[s(f,{appearance:"tertiary","icon-only":""},{default:t(()=>[s(r(T)),e(),ne]),_:1}),e()]),_:1}),e(),s(f,{to:{name:"diagnostics"},appearance:"tertiary","icon-only":"","data-testid":"nav-item-diagnostics"},{default:t(()=>[s(r(U)),e(),oe]),_:1})],!0)])]),e(),i("div",ae,[n.$slots.navigation?(p(),E("nav",se,[i("ul",null,[m(n.$slots,"navigation",{},void 0,!0)])])):g("",!0),e(),i("div",re,[i("div",ie,[m(n.$slots,"notifications",{},void 0,!0)]),e(),m(n.$slots,"notifications",{},()=>[r(_)("use state")?g("",!0):(p(),d(v,{key:0,class:"mb-4",appearance:"warning"},{default:t(()=>[i("ul",null,[i("li",{"data-testid":"warning-GLOBAL_STORE_TYPE_MEMORY",innerHTML:r(a)("common.warnings.GLOBAL_STORE_TYPE_MEMORY")},null,8,ce)])]),_:1}))],!0),e(),m(n.$slots,"default",{},void 0,!0)])])])}}}),pe=y(le,[["__scopeId","data-v-c98e0e53"]]),de=h({__name:"MeshNavigator",setup(c){return(o,_)=>{const a=l("RouterLink");return p(),d(k,{"data-testid":"meshes-navigator"},{default:t(()=>[s(a,{class:$({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="mesh-index-view")}),to:{name:"mesh-list-view"}},{default:t(()=>[e(`
      Meshes
    `)]),_:1},8,["class"])]),_:1})}}}),ue=h({__name:"ZoneEgressNavigator",setup(c){return(o,_)=>{const a=l("RouterLink");return p(),d(k,{"data-testid":"zone-egresses-navigator"},{default:t(()=>[s(a,{class:$({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="zone-egress-list-view")}),to:{name:"zone-egress-list-view"}},{default:t(()=>[e(`
      Zone Egresses
    `)]),_:1},8,["class"])]),_:1})}}}),_e=h({__name:"ZoneNavigator",setup(c){return(o,_)=>{const a=l("RouterLink");return p(),d(k,{"data-testid":"zones-navigator"},{default:t(()=>[s(a,{class:$({"is-active":[o.$route.name].concat(o.$route.matched.map(n=>n.name)).some(n=>n==="zone-index-view")}),to:{name:"zone-cp-list-view"}},{default:t(()=>[e(`
      Zones
    `)]),_:1},8,["class"])]),_:1})}}}),me=["alt"],he=h({__name:"App",setup(c){return(o,_)=>{const a=l("RouterView"),n=l("AppView"),L=l("RouteView"),b=l("DataSource");return p(),d(b,{src:"/control-plane/addresses"},{default:t(({data:f})=>[typeof f<"u"?(p(),d(L,{key:0,name:"app",attrs:{class:"kuma-ready"},"data-testid-root":"mesh-app"},{default:t(({t:v,can:w})=>[s(pe,{class:"kuma-application"},{home:t(()=>[i("img",{class:"logo",src:q,alt:`${v("common.product.name")} Logo`,"data-testid":"logo"},null,8,me)]),navigation:t(()=>[s(Z),e(),w("use zones")?(p(),d(_e,{key:0})):(p(),d(ue,{key:1})),e(),s(de)]),default:t(()=>[e(),e(),s(n,null,{default:t(()=>[s(a)]),_:1})]),_:2},1024)]),_:1})):g("",!0)]),_:1})}}}),ve=y(he,[["__scopeId","data-v-f821200e"]]);export{ve as default};
