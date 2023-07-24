import{j as _}from"./kongponents.es-406a7d3e.js";import{d as p,o as a,e as r,t,g as n,k as o,F as d,j as m,h as y,w as g,a as P,i as f,b as h,f as D}from"./index-abe682b3.js";import{z as C,n as N,B as O,D as b}from"./RouteView.vue_vue_type_script_setup_true_lang-99401e5a.js";const E=p({__name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}},setup(e){return(s,u)=>(a(),r("span",null,t(e.payload),1))}}),I=p({__name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(s,u)=>(a(),r("span",null,[n(`
    Envoy (`),o("strong",null,t(e.payload.envoy),1),n(") is unsupported by the current version of Kuma DP ("),o("strong",null,t(e.payload.kumaDp),1),n(") [Requirements: "),o("strong",null,t(e.payload.requirements),1),n(`].
  `)]))}}),k=p({__name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}},setup(e){return(s,u)=>(a(),r("span",null,[n(`
    Unsupported version of Kuma DP (`),o("strong",null,t(e.payload.kumaDp),1),n(`)
  `)]))}}),A=p({__name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(s,u)=>(a(),r("span",null,[n(`
    There is mismatch between versions of Zone CP (`),o("strong",null,t(e.payload.zoneCpVersion),1),n(`)
    and the Global CP (`),o("strong",null,t(e.payload.globalCpVersion),1),n(`)
  `)]))}}),V=p({__name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(s,u)=>(a(),r("span",null,[n(`
    There is a mismatch between versions of Kuma DP (`),o("strong",null,t(e.payload.kumaDp),1),n(`) and the Zone CP.
  `)]))}}),v={key:0,class:"stack"},j=p({__name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},setup(e){const s=e;function u(c=""){switch(c){case b:return I;case O:return k;case N:return V;case C:return A;default:return E}}return(c,B)=>s.warnings.length>0?(a(),r("div",v,[(a(!0),r(d,null,m(s.warnings,(l,i)=>(a(),r("div",{key:`${l.kind}/${i}`},[y(h(_),{appearance:"warning"},{alertMessage:g(()=>[(a(),P(f(u(l.kind)),{payload:l.payload},null,8,["payload"]))]),_:2},1024)]))),128))])):D("",!0)}});export{j as _};
