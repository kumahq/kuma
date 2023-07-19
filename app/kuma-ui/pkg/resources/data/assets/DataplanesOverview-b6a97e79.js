import{g as P}from"./kongponents.es-a3b26540.js";import{L as S}from"./LoadingBox-543d1594.js";import{O as T,a as F,b as I}from"./OnboardingPage-e9a8aa03.js";import{g as N,i as E,v as L,A as V,_ as C,f as H}from"./RouteView.vue_vue_type_script_setup_true_lang-071b2880.js";import{_ as M}from"./RouteTitle.vue_vue_type_script_setup_true_lang-0234a2f7.js";import{S as R}from"./StatusBadge-53125992.js";import{d as $,q as b,c as w,A as q,o as n,a as y,w as a,h as s,b as k,g as u,l as z,k as o,t as c,e as d,F as K}from"./index-54ac506b.js";const U={key:0,class:"status-loading-box mb-4"},W={key:1},j={class:"mb-4"},G=$({__name:"DataplanesOverview",setup(J){const p=N(),{t:x}=E(),D=[{label:"Mesh",key:"mesh"},{label:"Name",key:"name"},{label:"Status",key:"status"}],e=b({total:0,data:[]}),l=b(null),A=w(()=>e.value.data.length>0?"Success":"Waiting for DPPs"),m=w(()=>e.value.data.length>0?"The following data plane proxies (DPPs) are connected to the control plane:":null);q(function(){f()}),_();function f(){l.value!==null&&window.clearTimeout(l.value)}async function _(){let i=!1;const r=[];try{const{items:t}=await p.getAllDataplanes({size:10});if(Array.isArray(t))for(const B of t){const{name:v,mesh:g}=B,O=await p.getDataplaneOverviewFromMesh({mesh:g,name:v}),h=L(O.dataplaneInsight);h==="offline"&&(i=!0),r.push({status:h,name:v,mesh:g})}}catch(t){console.error(t)}e.value.data=r,e.value.total=e.value.data.length,i&&(f(),l.value=window.setTimeout(_,1e3))}return(i,r)=>(n(),y(C,null,{default:a(()=>[s(M,{title:k(x)("onboarding.routes.dataplanes-overview.title")},null,8,["title"]),u(),s(V,null,{default:a(()=>[s(T,null,{header:a(()=>[s(F,null,z({title:a(()=>[o("p",null,c(A.value),1)]),_:2},[m.value!==null?{name:"description",fn:a(()=>[o("p",null,c(m.value),1)]),key:"0"}:void 0]),1024)]),content:a(()=>[e.value.data.length===0?(n(),d("div",U,[s(S)])):(n(),d("div",W,[o("p",j,[o("b",null,"Found "+c(e.value.data.length)+" DPPs:",1)]),u(),s(k(P),{class:"mb-4",fetcher:()=>e.value,headers:D,"disable-pagination":""},{status:a(({rowValue:t})=>[t?(n(),y(R,{key:0,status:t},null,8,["status"])):(n(),d(K,{key:1},[u(`
                  —
                `)],64))]),_:1},8,["fetcher"])]))]),navigation:a(()=>[s(I,{"next-step":"onboarding-completed","previous-step":"onboarding-add-services-code","should-allow-next":e.value.data.length>0},null,8,["should-allow-next"])]),_:1})]),_:1})]),_:1}))}});const se=H(G,[["__scopeId","data-v-4588fbe4"]]);export{se as default};
