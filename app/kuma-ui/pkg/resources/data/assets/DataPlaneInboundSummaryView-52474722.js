import{d as I,k as S,S as g,a as l,o as c,b as p,w as t,e as _,m as r,f as m,t as f,l as y,c as k,C as x,v as R,x as C,a1 as N,_ as D}from"./index-0dcf85b4.js";import{_ as B}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-6ff90078.js";import{N as P}from"./NavTabs-e7b14354.js";const A=a=>(R("data-v-1fc90608"),a=a(),C(),a),T={class:"summary-title-wrapper"},$=A(()=>r("img",{"aria-hidden":"true",src:N},null,-1)),E={class:"summary-title"},j={key:1,class:"stack"},q=I({__name:"DataPlaneInboundSummaryView",props:{data:{}},setup(a){var w;const{t:i}=S(),h=g(),v=a,V=(((w=h.getRoutes().find(e=>e.name==="data-plane-inbound-summary-view"))==null?void 0:w.children)??[]).map(e=>{var n,s;const u=typeof e.name>"u"?(n=e.children)==null?void 0:n[0]:e,o=u.name,d=((s=u.meta)==null?void 0:s.module)??"";return{title:i(`data-planes.routes.item.navigation.${o}`),routeName:o,module:d}});return(e,u)=>{const o=l("RouterView"),d=l("AppView"),b=l("RouteView");return c(),p(b,{name:"data-plane-inbound-summary-view",params:{service:""}},{default:t(({route:n})=>[_(d,null,{title:t(()=>[r("div",T,[$,m(),r("h2",E,`
            Inbound :`+f(n.params.service),1)])]),default:t(()=>[m(),typeof v.data>"u"?(c(),p(B,{key:0},{message:t(()=>[r("p",null,f(y(i)("common.collection.summary.empty_message",{type:"Inbound"})),1)]),default:t(()=>[m(f(y(i)("common.collection.summary.empty_title",{type:"Inbound"}))+" ",1)]),_:1})):(c(),k("div",j,[_(P,{tabs:y(V)},null,8,["tabs"]),m(),_(o,null,{default:t(s=>[(c(),p(x(s.Component),{data:v.data},null,8,["data"]))]),_:1})]))]),_:2},1024)]),_:1})}}});const J=D(q,[["__scopeId","data-v-1fc90608"]]);export{J as default};
