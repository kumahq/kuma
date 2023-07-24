import{d as P,q as y,c as k,s as D,v as T,r as C,o as c,a as v,w as a,g as p,H as V,k as b,I as $,e as g,F as A,j as B,h as o,t as w,b as x,f as I}from"./index-abe682b3.js";import{_ as L}from"./StatusInfo.vue_vue_type_script_setup_true_lang-b76c0010.js";import{g as N,e as E,i as F,A as H,_ as M}from"./RouteView.vue_vue_type_script_setup_true_lang-99401e5a.js";import{_ as W}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-69dbcb95.js";import{T as j}from"./TabsWidget-39cc07d3.js";import{T as K}from"./TextWithCopyButton-a3aee751.js";import{_ as O}from"./RouteTitle.vue_vue_type_script_setup_true_lang-64e3d4a6.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-147f18f2.js";import"./kongponents.es-406a7d3e.js";import"./ErrorBlock-fe291eb6.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-c332ba11.js";import"./QueryParameter-70743f73.js";const R=b("h2",null,"Dataplanes",-1),U=P({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyPath:{type:String,required:!0},policyName:{type:String,required:!0}},setup(m){const e=m,u=N(),i=y(!1),s=y(!0),n=y(!1),r=y([]),t=y(""),d=k(()=>{const _=t.value.toLowerCase();return r.value.filter(({dataplane:l})=>l.name.toLowerCase().includes(_))});D(()=>e.policyName,function(){f()}),T(function(){f()});async function f(){n.value=!1,s.value=!0;try{const{items:_,total:l}=await u.getPolicyConnections({mesh:e.mesh,policyPath:e.policyPath,policyName:e.policyName});i.value=l>0,r.value=_??[]}catch{n.value=!0}finally{s.value=!1}}return(_,l)=>{const S=C("router-link");return c(),v(L,{"has-error":n.value,"is-loading":s.value,"is-empty":!i.value},{default:a(()=>[R,p(),V(b("input",{id:"dataplane-search","onUpdate:modelValue":l[0]||(l[0]=h=>t.value=h),type:"text",class:"k-input mt-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[$,t.value]]),p(),(c(!0),g(A,null,B(d.value,(h,q)=>(c(),g("p",{key:q,class:"mt-2","data-testid":"dataplane-name"},[o(S,{to:{name:"data-plane-detail-view",params:{mesh:h.dataplane.mesh,dataPlane:h.dataplane.name}}},{default:a(()=>[p(w(h.dataplane.name),1)]),_:2},1032,["to"])]))),128))]),_:1},8,["has-error","is-loading","is-empty"])}}}),z={class:"policy-details kcard-border"},G={class:"entity-heading","data-testid":"policy-single-entity"},J=P({__name:"PolicyDetails",props:{mesh:{type:String,required:!0},path:{type:String,required:!0},name:{type:String,required:!0},type:{type:String,required:!0}},setup(m){const e=m,u=N(),i=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],s=k(()=>({name:"policy-detail-view",params:{mesh:e.mesh,policy:e.name,policyPath:e.path}}));async function n(r){const{name:t,mesh:d,path:f}=e;return await u.getSinglePolicyEntity({name:t,mesh:d,path:f},r)}return(r,t)=>{const d=C("router-link");return c(),g("div",z,[o(j,{tabs:i},{tabHeader:a(()=>[b("h1",G,[p(w(e.type)+`:

          `,1),o(K,{text:e.name},{default:a(()=>[o(d,{to:s.value},{default:a(()=>[p(w(e.name),1)]),_:1},8,["to"])]),_:1},8,["text"])])]),overview:a(()=>[o(W,{id:"code-block-policy","resource-fetcher":n,"resource-fetcher-watch-key":e.name,"is-searchable":""},null,8,["resource-fetcher-watch-key"])]),"affected-dpps":a(()=>[o(U,{mesh:e.mesh,"policy-name":e.name,"policy-path":e.path},null,8,["mesh","policy-name","policy-path"])]),_:1})])}}}),le=P({__name:"PolicyDetailView",props:{mesh:{},policyPath:{},policyName:{}},setup(m){const e=m,u=E(),{t:i}=F(),s=k(()=>u.state.policyTypesByPath[e.policyPath]);return(n,r)=>(c(),v(M,null,{default:a(({route:t})=>[o(O,{title:x(i)("policies.routes.item.title",{name:t.params.policy})},null,8,["title"]),p(),o(H,{breadcrumbs:[{to:{name:"policies-list-view",params:{mesh:t.params.mesh,policyPath:t.params.policyPath}},text:x(i)("policies.routes.item.breadcrumbs")}]},{default:a(()=>[s.value?(c(),v(J,{key:0,name:e.policyName,mesh:e.mesh,path:e.policyPath,type:s.value.name},null,8,["name","mesh","path","type"])):I("",!0)]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{le as default};
