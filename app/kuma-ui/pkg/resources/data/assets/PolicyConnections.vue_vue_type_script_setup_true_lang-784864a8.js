import{d as C,r as t,c as k,v as x,k as N,x as P,o as i,e as q,w as d,h as c,a0 as D,a1 as B,i as f,j as h,q as L,g as S,t as V,b as F,F as b}from"./index-e1c5e7d3.js";import{_ as A}from"./StatusInfo.vue_vue_type_script_setup_true_lang-46e64e03.js";import{u as E}from"./index-6b720421.js";const I=f("h2",null,"Dataplanes",-1),K=C({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyPath:{type:String,required:!0},policyName:{type:String,required:!0}},setup(y){const s=y,v=E(),u=t(!1),o=t(!0),r=t(!1),p=t([]),l=t(""),_=k(()=>{const n=l.value.toLowerCase();return p.value.filter(({dataplane:e})=>e.name.toLowerCase().includes(n))});x(()=>s.policyName,function(){m()}),N(function(){m()});async function m(){r.value=!1,o.value=!0;try{const{items:n,total:e}=await v.getPolicyConnections({mesh:s.mesh,policyPath:s.policyPath,policyName:s.policyName});u.value=e>0,p.value=n??[]}catch{r.value=!0}finally{o.value=!1}}return(n,e)=>{const g=P("router-link");return i(),q(A,{"has-error":r.value,"is-loading":o.value,"is-empty":!u.value},{default:d(()=>[I,c(),D(f("input",{id:"dataplane-search","onUpdate:modelValue":e[0]||(e[0]=a=>l.value=a),type:"text",class:"k-input mt-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[B,l.value]]),c(),(i(!0),h(b,null,L(F(_),(a,w)=>(i(),h("p",{key:w,class:"mt-2","data-testid":"dataplane-name"},[S(g,{to:{name:"data-plane-detail-view",params:{mesh:a.dataplane.mesh,dataPlane:a.dataplane.name}}},{default:d(()=>[c(V(a.dataplane.name),1)]),_:2},1032,["to"])]))),128))]),_:1},8,["has-error","is-loading","is-empty"])}}});export{K as _};
