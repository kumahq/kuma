import{d as I,k as g,T as x,a as l,o as m,b as p,w as t,e as _,m as r,f as c,t as y,l as f,c as R,F as S,B as k,v as N,x as B,a1 as C,_ as D}from"./index-3733fdc1.js";import{_ as P}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-7a730a6c.js";import{N as T}from"./NavTabs-5a31cf31.js";const A=a=>(N("data-v-708e12d6"),a=a(),B(),a),F={class:"summary-title-wrapper"},$=A(()=>r("img",{"aria-hidden":"true",src:C},null,-1)),E={class:"summary-title"},j=I({__name:"DataPlaneInboundSummaryView",props:{data:{}},setup(a){var v;const{t:u}=g(),V=x(),w=a,h=(((v=V.getRoutes().find(e=>e.name==="data-plane-inbound-summary-view"))==null?void 0:v.children)??[]).map(e=>{var o,s;const d=typeof e.name>"u"?(o=e.children)==null?void 0:o[0]:e,n=d.name,i=((s=d.meta)==null?void 0:s.module)??"";return{title:u(`data-planes.routes.item.navigation.${n}`),routeName:n,module:i}});return(e,d)=>{const n=l("RouterView"),i=l("AppView"),b=l("RouteView");return m(),p(b,{name:"data-plane-inbound-summary-view",params:{service:""}},{default:t(({route:o})=>[_(i,null,{title:t(()=>[r("div",F,[$,c(),r("h2",E,`
            Inbound :`+y(o.params.service),1)])]),default:t(()=>[c(),typeof w.data>"u"?(m(),p(P,{key:0},{message:t(()=>[r("p",null,y(f(u)("common.collection.summary.empty_message",{type:"Inbound"})),1)]),default:t(()=>[c(y(f(u)("common.collection.summary.empty_title",{type:"Inbound"}))+" ",1)]),_:1})):(m(),R(S,{key:1},[_(T,{tabs:f(h)},null,8,["tabs"]),c(),_(n,null,{default:t(s=>[(m(),p(k(s.Component),{data:w.data},null,8,["data"]))]),_:1})],64))]),_:2},1024)]),_:1})}}});const J=D(j,[["__scopeId","data-v-708e12d6"]]);export{J as default};
