import{a as ae,u as te,k as K,i as le,p as se}from"./production-7bbeb92b.js";import{m as ne,I as J,P as oe,J as re}from"./kongponents.es-130f96bb.js";import{Q}from"./QueryParameter-70743f73.js";import{u as ie}from"./store-d9c38acd.js";import{D as ue}from"./DataOverview-e6a9bbf4.js";import{d as z,o as m,a as A,w as r,e as g,u as a,f as i,r as t,c as F,q as R,i as ce,s as pe,h as w,g as n,X as me,a5 as de,F as W,l as G,t as I,b as C,I as H,p as ye,k as ve}from"./runtime-dom.esm-bundler-062436f2.js";import{F as fe}from"./FrameSkeleton-469f3e69.js";import{_ as X}from"./LabelList.vue_vue_type_style_index_0_lang-2994b5e0.js";import{u as Y,a as he}from"./index-0228babc.js";import{T as _e}from"./TabsWidget-d7b7af6a.js";import{_ as ge}from"./YamlView.vue_vue_type_script_setup_true_lang-509c3a62.js";import{_ as be}from"./_plugin-vue_export-helper-c27b6911.js";import"./datadogLogEvents-302eea7b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-b391650a.js";import"./ErrorBlock-863727f0.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-77c2a6ee.js";import"./StatusBadge-5a5c160e.js";import"./TagList-ab65e41f.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-3cb2154f.js";import"./toYaml-4e00099e.js";const Pe=z({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(d){const l=d;return(E,S)=>(m(),A(a(J),{class:"docs-link",appearance:"outline",target:"_blank",to:l.href},{default:r(()=>[g(a(ne),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),i(`

    Documentation
  `)]),_:1},8,["to"]))}}),ke=n("h4",null,"Dataplanes",-1),we=z({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},setup(d){const l=d,E=Y(),S=t(!1),N=t(!0),q=t(!1),y=t([]),P=t(""),b=F(()=>{const u=P.value.toLowerCase();return y.value.filter(({dataplane:s})=>s.name.toLowerCase().includes(u))});R(()=>l.policyName,function(){v()}),ce(function(){v()});async function v(){q.value=!1,N.value=!0;try{const{items:u,total:s}=await E.getPolicyConnections({mesh:l.mesh,policyType:l.policyType,policyName:l.policyName});S.value=s>0,y.value=u??[]}catch{q.value=!0}finally{N.value=!1}}return(u,s)=>{const f=pe("router-link");return m(),w("div",null,[g(X,{"has-error":q.value,"is-loading":N.value,"is-empty":!S.value},{default:r(()=>[n("ul",null,[n("li",null,[ke,i(),me(n("input",{id:"dataplane-search","onUpdate:modelValue":s[0]||(s[0]=c=>P.value=c),type:"text",class:"k-input mb-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[de,P.value]]),i(),(m(!0),w(W,null,G(a(b),(c,k)=>(m(),w("p",{key:k,class:"my-1","data-testid":"dataplane-name"},[g(f,{to:{name:"data-plane-detail-view",params:{mesh:c.dataplane.mesh,dataPlane:c.dataplane.name}}},{default:r(()=>[i(I(c.dataplane.name),1)]),_:2},1032,["to"])]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])])}}}),Ee=d=>(ye("data-v-e43a7936"),d=d(),ve(),d),Se={key:0,class:"mb-4"},De=Ee(()=>n("p",null,[n("strong",null,"Warning"),i(` This policy is experimental. If you encountered any problem please open an
            `),n("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),xe={class:"entity-heading","data-testid":"policy-single-entity"},Te={"data-testid":"policy-overview-tab"},Ce={class:"config-wrapper"},Ie=z({__name:"PolicyView",props:{selectedPolicyName:{type:String,required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(d){const l=d,E=Y(),S=he(),N=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],q=ae(),y=te(),P=ie(),b=t(!0),v=t(!1),u=t(null),s=t(!0),f=t(!1),c=t(!1),k=t(!1),D=t({}),x=t(null),M=t(null),B=t(l.offset),U=t({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),h=F(()=>P.state.policyTypesByPath[l.policyPath]),Z=F(()=>P.state.policyTypes.map(e=>({length:P.state.sidebar.insights.mesh.policies[e.name]??0,label:e.name,value:e.path,selected:e.path===l.policyPath})));R(()=>y.params.mesh,function(){y.name===l.policyPath&&(b.value=!0,v.value=!1,s.value=!0,f.value=!1,c.value=!1,k.value=!1,u.value=null,L(0))}),R(()=>y.query.ns,function(){b.value=!0,v.value=!1,s.value=!0,f.value=!1,c.value=!1,k.value=!1,u.value=null,L(0)}),L(l.offset);async function L(e){B.value=e,Q.set("offset",e>0?e:null),b.value=!0,u.value=null;const p=y.query.ns||null,o=y.params.mesh,T=h.value.path;try{let _;if(o!==null&&p!==null)_=[await E.getSinglePolicyEntity({mesh:o,path:T,name:p})],M.value=null;else{const $={size:K,offset:e},V=await E.getAllPolicyEntitiesFromMesh({mesh:o,path:T},$);_=V.items??[],M.value=V.next}if(_.length>0){U.value.data=_.map(V=>ee(V)),k.value=!1,v.value=!1;const $=l.selectedPolicyName??_[0].name;await O({name:$,mesh:o,path:T})}else U.value.data=[],k.value=!0,v.value=!0,f.value=!0}catch(_){_ instanceof Error?u.value=_:console.error(_),v.value=!0}finally{b.value=!1,s.value=!1}}function j(e){q.push({name:"policy",params:{...y.params,policyPath:e.value}})}function ee(e){if(!e.mesh)return e;const p=e,o={name:"mesh-detail-view",params:{mesh:e.mesh}};return p.meshRoute=o,p}async function O(e){c.value=!1,s.value=!0,f.value=!1;try{const p=await E.getSinglePolicyEntity({mesh:e.mesh,path:h.value.path,name:e.name});if(p){const o=["type","name","mesh"];D.value=le(p,o),Q.set("policy",D.value.name),x.value=se(p)}else D.value={},f.value=!0}catch(p){c.value=!0,console.error(p)}finally{s.value=!1}}return(e,p)=>a(h)?(m(),w("div",{key:0,class:H(["relative",a(h).path])},[a(h).isExperimental?(m(),w("div",Se,[g(a(oe),{appearance:"warning"},{alertMessage:r(()=>[De]),_:1})])):C("",!0),i(),g(fe,null,{default:r(()=>[g(ue,{"selected-entity-name":D.value.name,"page-size":a(K),error:u.value,"is-loading":b.value,"empty-state":{title:"No Data",message:`There are no ${a(h).name} policies present.`},"table-data":U.value,"table-data-is-empty":k.value,next:M.value,"page-offset":B.value,onTableAction:O,onLoadData:L},{additionalControls:r(()=>[g(a(re),{label:"Policies",items:a(Z),"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:j},{"item-template":r(({item:o})=>[n("span",{class:H({"policy-type-empty":o.length===0})},I(o.label),3)]),_:1},8,["items"]),i(),g(Pe,{href:`${a(S)("KUMA_DOCS_URL")}/policies/${a(h).path}/?${a(S)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"]),i(),e.$route.query.ns?(m(),A(a(J),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"policy",params:{policyPath:l.policyPath}}},{default:r(()=>[i(`
            View all
          `)]),_:1},8,["to"])):C("",!0)]),default:r(()=>[i(`
        >
        `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"]),i(),v.value===!1?(m(),A(_e,{key:0,"has-error":u.value!==null,error:u.value,"is-loading":b.value,tabs:N},{tabHeader:r(()=>[n("h1",xe,I(a(h).name)+": "+I(D.value.name),1)]),overview:r(()=>[g(X,{"has-error":c.value,"is-loading":s.value,"is-empty":f.value},{default:r(()=>[n("div",Te,[n("ul",null,[(m(!0),w(W,null,G(D.value,(o,T)=>(m(),w("li",{key:T},[n("h4",null,I(T),1),i(),n("p",null,I(o),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),i(),n("div",Ce,[x.value!==null?(m(),A(ge,{key:0,id:"code-block-policy","has-error":c.value,"is-loading":s.value,"is-empty":f.value,content:x.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):C("",!0)])]),"affected-dpps":r(()=>[x.value!==null?(m(),A(we,{key:0,mesh:x.value.mesh,"policy-name":x.value.name,"policy-type":a(h).path},null,8,["mesh","policy-name","policy-type"])):C("",!0)]),_:1},8,["has-error","error","is-loading"])):C("",!0)]),_:1})],2)):C("",!0)}});const Ye=be(Ie,[["__scopeId","data-v-e43a7936"]]);export{Ye as default};
