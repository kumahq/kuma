import{d as l,o as a,c as r,y as t,e as n,g as s,F as _,z as d,a as m,w as y,k as g,a2 as P,u as O,X as D,I as f,a1 as C,a3 as N,a4 as b}from"./index-1f1e085f.js";const h=l({__name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}},setup(e){return(o,u)=>(a(),r("span",null,t(e.payload),1))}}),E=l({__name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),r("span",null,[n(`
    Envoy (`),s("strong",null,t(e.payload.envoy),1),n(") is unsupported by the current version of Kuma DP ("),s("strong",null,t(e.payload.kumaDp),1),n(") [Requirements: "),s("strong",null,t(e.payload.requirements),1),n(`].
  `)]))}}),I=l({__name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),r("span",null,[n(`
    Unsupported version of Kuma DP (`),s("strong",null,t(e.payload.kumaDp),1),n(`)
  `)]))}}),A=l({__name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),r("span",null,[n(`
    There is mismatch between versions of Zone CP (`),s("strong",null,t(e.payload.zoneCpVersion),1),n(`)
    and the Global CP (`),s("strong",null,t(e.payload.globalCpVersion),1),n(`)
  `)]))}}),V=l({__name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}},setup(e){return(o,u)=>(a(),r("span",null,[n(`
    There is a mismatch between versions of Kuma DP (`),s("strong",null,t(e.payload.kumaDp),1),n(`) and the Zone CP.
  `)]))}}),x=l({__name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},setup(e){const o=e;function u(c=""){switch(c){case b:return E;case N:return I;case C:return V;case f:return A;default:return h}}return(c,k)=>(a(),r("ul",null,[(a(!0),r(_,null,d(o.warnings,(p,i)=>(a(),r("li",{key:`${p.kind}/${i}`,class:"mb-1"},[m(O(D),{appearance:"warning"},{alertMessage:y(()=>[(a(),g(P(u(p.kind)),{payload:p.payload},null,8,["payload"]))]),_:2},1024)]))),128))]))}});export{x as _};
