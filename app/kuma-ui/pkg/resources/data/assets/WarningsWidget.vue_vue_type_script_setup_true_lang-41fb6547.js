import{O as y,_ as g}from"./kongponents.es-8abed680.js";import{d as l,o as a,h as s,t as r,f as n,g as t,a as i,w as _,m as P,e as f,y as O,u as m,F as b}from"./runtime-dom.esm-bundler-a6f4ece5.js";import{I as D,l as h,m as C,n as N}from"./production-0f1ffdb6.js";const E=l({__name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}},setup(e){return(o,u)=>(a(),s("span",null,r(e.payload),1))}}),I=l({__name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),s("span",null,[n(`
    Envoy (`),t("strong",null,r(e.payload.envoy),1),n(") is unsupported by the current version of Kuma DP ("),t("strong",null,r(e.payload.kumaDp),1),n(") [Requirements: "),t("strong",null,r(e.payload.requirements),1),n(`].
  `)]))}}),A=l({__name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),s("span",null,[n(`
    There is a mismatch between versions of Kuma DP (`),t("strong",null,r(e.payload.kumaDp),1),n(`) and the Zone CP.
  `)]))}}),V=l({__name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),s("span",null,[n(`
    Unsupported version of Kuma DP (`),t("strong",null,r(e.payload.kumaDp),1),n(`)
  `)]))}}),B=l({__name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),s("span",null,[n(`
    There is mismatch between versions of Zone CP (`),t("strong",null,r(e.payload.zoneCpVersion),1),n(`)
    and the Global CP (`),t("strong",null,r(e.payload.globalCpVersion),1),n(`)
  `)]))}}),T=l({__name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},setup(e){const o=e;function u(c=""){switch(c){case N:return I;case C:return V;case h:return A;case D:return B;default:return E}}return(c,S)=>(a(),i(m(g),{"border-variant":"noBorder"},{body:_(()=>[t("ul",null,[(a(!0),s(b,null,P(o.warnings,(p,d)=>(a(),s("li",{key:`${p.kind}/${d}`,class:"mb-1"},[f(m(y),{appearance:"warning"},{alertMessage:_(()=>[(a(),i(O(u(p.kind)),{payload:p.payload},null,8,["payload"]))]),_:2},1024)]))),128))])]),_:1}))}});export{T as _};
