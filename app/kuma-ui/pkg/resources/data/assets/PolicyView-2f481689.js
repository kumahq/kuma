import{b as ee,u as ae}from"./vue-router-67937a96.js";import{m as te,P as W,T as le,U as se}from"./kongponents.es-7ead79da.js";import{h as ne,n as oe}from"./production-8efaeab1.js";import{k as $}from"./kumaApi-08f7fc23.js";import{b as Q}from"./constants-31fdaf55.js";import{Q as H}from"./QueryParameter-70743f73.js";import{u as re}from"./store-ec4aec64.js";import{D as ie}from"./DataOverview-6ea76f5e.js";import{d as z,o as m,a as L,w as r,e as g,u as a,f as i,r as t,c as F,s as R,k as ue,i as ce,h as k,g as n,S as pe,a5 as me,F as j,m as G,t as N,b as C,q as K,p as de,j as ye}from"./runtime-dom.esm-bundler-fd3ecc5a.js";import{F as ve}from"./FrameSkeleton-c95e2670.js";import{_ as Y}from"./LabelList.vue_vue_type_style_index_0_lang-2c60d3f9.js";import{T as fe}from"./TabsWidget-8139dab1.js";import{_ as he}from"./YamlView.vue_vue_type_script_setup_true_lang-545eb1cb.js";import{u as _e}from"./index-be4d4b11.js";import{_ as ge}from"./_plugin-vue_export-helper-c27b6911.js";import"./vuex.esm-bundler-4e6e06ec.js";import"./datadogLogEvents-302eea7b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-7c04ed58.js";import"./ErrorBlock-04954e92.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-ef489e58.js";import"./StatusBadge-760d0ebe.js";import"./TagList-d36d0a47.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2bd08d1e.js";import"./_commonjsHelpers-edff4021.js";import"./toYaml-4e00099e.js";const be=z({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(d){const l=d;return(S,E)=>(m(),L(a(W),{class:"docs-link",appearance:"outline",target:"_blank",to:l.href},{default:r(()=>[g(a(te),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),i(`

    Documentation
  `)]),_:1},8,["to"]))}}),Pe=n("h4",null,"Dataplanes",-1),we=z({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},setup(d){const l=d,S=t(!1),E=t(!0),q=t(!1),y=t([]),P=t(""),b=F(()=>{const u=P.value.toLowerCase();return y.value.filter(({dataplane:s})=>s.name.toLowerCase().includes(u))});R(()=>l.policyName,function(){v()}),ue(function(){v()});async function v(){q.value=!1,E.value=!0;try{const{items:u,total:s}=await $.getPolicyConnections({mesh:l.mesh,policyType:l.policyType,policyName:l.policyName});S.value=s>0,y.value=u??[]}catch{q.value=!0}finally{E.value=!1}}return(u,s)=>{const f=ce("router-link");return m(),k("div",null,[g(Y,{"has-error":q.value,"is-loading":E.value,"is-empty":!S.value},{default:r(()=>[n("ul",null,[n("li",null,[Pe,i(),pe(n("input",{id:"dataplane-search","onUpdate:modelValue":s[0]||(s[0]=c=>P.value=c),type:"text",class:"k-input mb-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[me,P.value]]),i(),(m(!0),k(j,null,G(a(b),(c,w)=>(m(),k("p",{key:w,class:"my-1","data-testid":"dataplane-name"},[g(f,{to:{name:"data-plane-detail-view",params:{mesh:c.dataplane.mesh,dataPlane:c.dataplane.name}}},{default:r(()=>[i(N(c.dataplane.name),1)]),_:2},1032,["to"])]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])])}}}),ke=d=>(de("data-v-9387bba4"),d=d(),ye(),d),Se={key:0,class:"mb-4"},Ee=ke(()=>n("p",null,[n("strong",null,"Warning"),i(` This policy is experimental. If you encountered any problem please open an
            `),n("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),De={class:"entity-heading","data-testid":"policy-single-entity"},Te={"data-testid":"policy-overview-tab"},xe={class:"config-wrapper"},Ce=z({__name:"PolicyView",props:{selectedPolicyName:{type:String,required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(d){const l=d,S=_e(),E=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],q=ee(),y=ae(),P=re(),b=t(!0),v=t(!1),u=t(null),s=t(!0),f=t(!1),c=t(!1),w=t(!1),D=t({}),T=t(null),V=t(null),B=t(l.offset),M=t({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),h=F(()=>P.state.policyTypesByPath[l.policyPath]),Z=F(()=>P.state.policyTypes.map(e=>({length:P.state.sidebar.insights.mesh.policies[e.name]??0,label:e.name,value:e.path,selected:e.path===l.policyPath})));R(()=>y.params.mesh,function(){y.name===l.policyPath&&(b.value=!0,v.value=!1,s.value=!0,f.value=!1,c.value=!1,w.value=!1,u.value=null,A(0))}),R(()=>y.query.ns,function(){b.value=!0,v.value=!1,s.value=!0,f.value=!1,c.value=!1,w.value=!1,u.value=null,A(0)}),A(l.offset);async function A(e){B.value=e,H.set("offset",e>0?e:null),b.value=!0,u.value=null;const p=y.query.ns||null,o=y.params.mesh,x=h.value.path;try{let _;if(o!==null&&p!==null)_=[await $.getSinglePolicyEntity({mesh:o,path:x,name:p})],V.value=null;else{const I={size:Q,offset:e},U=await $.getAllPolicyEntitiesFromMesh({mesh:o,path:x},I);_=U.items??[],V.value=U.next}if(_.length>0){M.value.data=_.map(U=>X(U)),w.value=!1,v.value=!1;const I=l.selectedPolicyName??_[0].name;await O({name:I,mesh:o,path:x})}else M.value.data=[],w.value=!0,v.value=!0,f.value=!0}catch(_){_ instanceof Error?u.value=_:console.error(_),v.value=!0}finally{b.value=!1,s.value=!1}}function J(e){q.push({name:"policy",params:{...y.params,policyPath:e.value}})}function X(e){if(!e.mesh)return e;const p=e,o={name:"mesh-detail-view",params:{mesh:e.mesh}};return p.meshRoute=o,p}async function O(e){c.value=!1,s.value=!0,f.value=!1;try{const p=await $.getSinglePolicyEntity({mesh:e.mesh,path:h.value.path,name:e.name});if(p){const o=["type","name","mesh"];D.value=ne(p,o),H.set("policy",D.value.name),T.value=oe(p)}else D.value={},f.value=!0}catch(p){c.value=!0,console.error(p)}finally{s.value=!1}}return(e,p)=>a(h)?(m(),k("div",{key:0,class:K(["relative",a(h).path])},[a(h).isExperimental?(m(),k("div",Se,[g(a(le),{appearance:"warning"},{alertMessage:r(()=>[Ee]),_:1})])):C("",!0),i(),g(ve,null,{default:r(()=>[g(ie,{"selected-entity-name":D.value.name,"page-size":a(Q),error:u.value,"is-loading":b.value,"empty-state":{title:"No Data",message:`There are no ${a(h).name} policies present.`},"table-data":M.value,"table-data-is-empty":w.value,next:V.value,"page-offset":B.value,onTableAction:O,onLoadData:A},{additionalControls:r(()=>[g(a(se),{label:"Policies",items:a(Z),"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:J},{"item-template":r(({item:o})=>[n("span",{class:K({"policy-type-empty":o.length===0})},N(o.label),3)]),_:1},8,["items"]),i(),g(be,{href:`${a(S)("KUMA_DOCS_URL")}/policies/${a(h).path}/?${a(S)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"]),i(),e.$route.query.ns?(m(),L(a(W),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"policy",params:{policyPath:l.policyPath}}},{default:r(()=>[i(`
            View all
          `)]),_:1},8,["to"])):C("",!0)]),default:r(()=>[i(`
        >
        `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"]),i(),v.value===!1?(m(),L(fe,{key:0,"has-error":u.value!==null,error:u.value,"is-loading":b.value,tabs:E},{tabHeader:r(()=>[n("h1",De,N(a(h).name)+": "+N(D.value.name),1)]),overview:r(()=>[g(Y,{"has-error":c.value,"is-loading":s.value,"is-empty":f.value},{default:r(()=>[n("div",Te,[n("ul",null,[(m(!0),k(j,null,G(D.value,(o,x)=>(m(),k("li",{key:x},[n("h4",null,N(x),1),i(),n("p",null,N(o),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),i(),n("div",xe,[T.value!==null?(m(),L(he,{key:0,id:"code-block-policy","has-error":c.value,"is-loading":s.value,"is-empty":f.value,content:T.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):C("",!0)])]),"affected-dpps":r(()=>[T.value!==null?(m(),L(we,{key:0,mesh:T.value.mesh,"policy-name":T.value.name,"policy-type":a(h).path},null,8,["mesh","policy-name","policy-type"])):C("",!0)]),_:1},8,["has-error","error","is-loading"])):C("",!0)]),_:1})],2)):C("",!0)}});const aa=ge(Ce,[["__scopeId","data-v-9387bba4"]]);export{aa as default};
