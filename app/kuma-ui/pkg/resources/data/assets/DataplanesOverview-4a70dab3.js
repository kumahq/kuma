import{d as B,R as S,L as N,t as b,f as w,W as T,o as n,g as y,w as a,h as s,i as D,C as F,l as u,X as I,m as l,D as c,j as d,U as E,Y as L,F as C,A as R,_ as V,q as H}from"./index-18fd9432.js";import{L as M}from"./LoadingBox-4457392c.js";import{O as U,a as W,b as $}from"./OnboardingPage-d7162bb4.js";import{g as j}from"./dataplane-30467516.js";import"./store-a3cd7f7b.js";const q={key:0,class:"status-loading-box mb-4"},z={key:1},K={class:"mb-4"},X=B({__name:"DataplanesOverview",setup(Y){const p=S(),{t:k}=N(),x=[{label:"Mesh",key:"mesh"},{label:"Name",key:"name"},{label:"Status",key:"status"}],e=b({total:0,data:[]}),o=b(null),A=w(()=>e.value.data.length>0?"Success":"Waiting for DPPs"),m=w(()=>e.value.data.length>0?"The following data plane proxies (DPPs) are connected to the control plane:":null);T(function(){_()}),v();function _(){o.value!==null&&window.clearTimeout(o.value)}async function v(){let i=!1;const r=[];try{const{items:t}=await p.getAllDataplanes({size:10});if(Array.isArray(t))for(const O of t){const{name:f,mesh:g}=O,P=await p.getDataplaneOverviewFromMesh({mesh:g,name:f}),h=j(P.dataplaneInsight);h==="offline"&&(i=!0),r.push({status:h,name:f,mesh:g})}}catch(t){console.error(t)}e.value.data=r,e.value.total=e.value.data.length,i&&(_(),o.value=window.setTimeout(v,1e3))}return(i,r)=>(n(),y(V,null,{default:a(()=>[s(F,{title:D(k)("onboarding.routes.dataplanes-overview.title")},null,8,["title"]),u(),s(R,null,{default:a(()=>[s(U,null,{header:a(()=>[s(W,null,I({title:a(()=>[l("p",null,c(A.value),1)]),_:2},[m.value!==null?{name:"description",fn:a(()=>[l("p",null,c(m.value),1)]),key:"0"}:void 0]),1024)]),content:a(()=>[e.value.data.length===0?(n(),d("div",q,[s(M)])):(n(),d("div",z,[l("p",K,[l("b",null,"Found "+c(e.value.data.length)+" DPPs:",1)]),u(),s(D(E),{class:"mb-4",fetcher:()=>e.value,headers:x,"disable-pagination":""},{status:a(({rowValue:t})=>[t?(n(),y(L,{key:0,status:t},null,8,["status"])):(n(),d(C,{key:1},[u(`
                  —
                `)],64))]),_:1},8,["fetcher"])]))]),navigation:a(()=>[s($,{"next-step":"onboarding-completed","previous-step":"onboarding-add-services-code","should-allow-next":e.value.data.length>0},null,8,["should-allow-next"])]),_:1})]),_:1})]),_:1}))}});const ae=H(X,[["__scopeId","data-v-4588fbe4"]]);export{ae as default};
