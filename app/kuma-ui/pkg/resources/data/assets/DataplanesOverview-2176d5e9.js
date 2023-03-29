import{d as S,l as A,r as b,j as w,aa as B,s as T,o as s,k as y,w as t,a as n,Q as F,u as o,g as l,y as u,c as d,e as k,F as N,S as E,H as I}from"./index-e096fb01.js";import{L as H}from"./LoadingBox-68c98c6a.js";import{O as L,a as C,b as M}from"./OnboardingPage-c6c18443.js";import{S as R}from"./StatusBadge-79f7109b.js";const V={key:0,class:"status-loading-box mb-4"},j={key:1},z={class:"mb-4"},K=S({__name:"DataplanesOverview",setup(Q){const p=A(),x=[{label:"Mesh",key:"mesh"},{label:"Name",key:"name"},{label:"Status",key:"status"}],e=b({total:0,data:[]}),i=b(null),D=w(()=>e.value.data.length>0?"Success":"Waiting for DPPs"),m=w(()=>e.value.data.length>0?"The following data plane proxies (DPPs) are connected to the control plane:":null);B(function(){v()}),g();function v(){i.value!==null&&window.clearTimeout(i.value)}async function g(){let r=!1;const c=[];try{const{items:a}=await p.getAllDataplanes({size:10});if(Array.isArray(a))for(const O of a){const{name:_,mesh:f}=O,P=await p.getDataplaneOverviewFromMesh({mesh:f,name:_}),h=T(P.dataplaneInsight);h==="offline"&&(r=!0),c.push({status:h,name:_,mesh:f})}}catch(a){console.error(a)}e.value.data=c,e.value.total=e.value.data.length,r&&(v(),i.value=window.setTimeout(g,1e3))}return(r,c)=>(s(),y(M,null,{header:t(()=>[n(L,null,F({title:t(()=>[l("p",null,u(o(D)),1)]),_:2},[o(m)!==null?{name:"description",fn:t(()=>[l("p",null,u(o(m)),1)]),key:"0"}:void 0]),1024)]),content:t(()=>[e.value.data.length===0?(s(),d("div",V,[n(H)])):(s(),d("div",j,[l("p",z,[l("b",null,"Found "+u(e.value.data.length)+" DPPs:",1)]),k(),n(o(E),{class:"mb-4",fetcher:()=>e.value,headers:x,"disable-pagination":""},{status:t(({rowValue:a})=>[a?(s(),y(R,{key:0,status:a},null,8,["status"])):(s(),d(N,{key:1},[k(`
              —
            `)],64))]),_:1},8,["fetcher"])]))]),navigation:t(()=>[n(C,{"next-step":"onboarding-completed","previous-step":"onboarding-add-services-code","should-allow-next":e.value.data.length>0},null,8,["should-allow-next"])]),_:1}))}});const J=I(K,[["__scopeId","data-v-9ed5a755"]]);export{J as default};
