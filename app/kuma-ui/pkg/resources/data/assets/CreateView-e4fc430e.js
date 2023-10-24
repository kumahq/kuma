import{d as A,y as f,z as de,U as me,r as z,o as r,l as d,p as o,j as l,w as a,i as k,k as e,aA as G,aB as H,aC as pe,aD as _e,aE as fe,aF as j,n,S as V,g as x,e as Q,h as w,H as s,Q as ve,a4 as be,aG as U,F as L,m as C,I as ze,X as he,K as X,aH as ge,t as ke}from"./index-9e09c995.js";import{_ as $}from"./CodeBlock.vue_vue_type_style_index_0_lang-5aea095c.js";import{g as ye}from"./dataplane-0a086c06.js";const Ee=["data-test-state"],Ce={class:"scanner-content"},$e={class:"mr-1"},Ie=A({__name:"EntityScanner",props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],setup(y,{emit:t}){const v=y,_=t,I=f(0),h=f(!1),c=f(!1),E=f(null);de(function(){S()}),me(function(){m()});function S(){h.value=!0,c.value=!1,m(),E.value=window.setInterval(async()=>{I.value++,await v.loaderFunction(),(I.value===v.retries||v.canComplete===!0)&&(m(),h.value=!1,c.value=!0,_("hide-siblings",!0))},v.interval)}function m(){E.value!==null&&window.clearInterval(E.value)}return(g,N)=>{const T=z("KEmptyState");return r(),d("div",{class:"scanner","data-test-state":h.value?"waiting":y.hasError?"error":"success"},[o("div",Ce,[l(T,{"cta-is-hidden":""},{title:a(()=>[o("span",$e,[h.value?(r(),k(e(G),{key:0,color:e(H)},null,8,["color"])):y.hasError?(r(),k(e(pe),{key:1,color:e(_e)},null,8,["color"])):(r(),k(e(fe),{key:2,color:e(j)},null,8,["color"]))]),n(),h.value?V(g.$slots,"loading-title",{key:0}):y.hasError?V(g.$slots,"error-title",{key:1}):V(g.$slots,"complete-title",{key:2})]),message:a(()=>[h.value?V(g.$slots,"loading-content",{key:0}):y.hasError?V(g.$slots,"error-content",{key:1}):V(g.$slots,"complete-content",{key:2})]),_:3})])],8,Ee)}}}),we={class:"form-step-title"},Ke=o("span",{class:"form-step-number"},"1",-1),Ve={class:"instruction-list"},Se={href:"https://helm.sh/docs/intro/install/"},Ne={class:"form-step-title"},Te=o("span",{class:"form-step-number"},"2",-1),Ze=o("p",null,"On your local machine, create a namespace in your Kubernetes cluster and pull down the kong Helm repo.",-1),qe={class:"instruction-list"},Le={class:"form-step-title"},Ae=o("span",{class:"form-step-number"},"3",-1),De={class:"form-step-title"},Be=o("span",{class:"form-step-number"},"4",-1),Re={class:"field-group-label mt-4"},Ue={class:"mt-4"},xe=A({__name:"ZoneCreateKubernetesInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},zoneIngressEnabled:{type:Boolean,required:!0},zoneEgressEnabled:{type:Boolean,required:!0},token:{type:String,required:!0},base64EncodedToken:{type:String,required:!0}},setup(y){const t=x(),v=Q(),_=y,I=w(()=>t.t("zones.form.kubernetes.secret.createSecretCommand",{token:_.base64EncodedToken}).trim()),h=w(()=>{const c={zoneName:_.zoneName,globalKdsAddress:_.globalKdsAddress,zoneIngressEnabled:String(_.zoneIngressEnabled),zoneEgressEnabled:String(_.zoneEgressEnabled)};return typeof v.params.virtualControlPlaneId=="string"&&(c.controlPlaneId=v.params.virtualControlPlaneId),t.t("zones.form.kubernetes.connectZone.config",c).trim()});return(c,E)=>(r(),d("div",null,[o("h3",we,[Ke,n(" "+s(e(t).t("zones.form.kubernetes.prerequisites.title")),1)]),n(),o("ul",Ve,[o("li",null,[o("b",null,s(e(t).t("zones.form.kubernetes.prerequisites.step1Label"))+s(_.zoneIngressEnabled?" "+e(t).t("zones.form.kubernetes.prerequisites.step1LabelAddendum"):""),1),n(`:
        `+s(e(t).t("zones.form.kubernetes.prerequisites.step1Description",{productName:e(t).t("common.product.name")})),1)]),n(),o("li",null,[o("b",null,s(e(t).t("zones.form.kubernetes.prerequisites.step2Label")),1),n(`:
        `+s(e(t).t("zones.form.kubernetes.prerequisites.step2Description")),1)]),n(),o("li",null,[o("a",Se,s(e(t).t("zones.form.kubernetes.prerequisites.step3LinkTitle")),1),n(" "+s(e(t).t("zones.form.kubernetes.prerequisites.step3Tail")),1)])]),n(),o("h3",Ne,[Te,n(" "+s(e(t).t("zones.form.kubernetes.helm.title")),1)]),n(),Ze,n(),o("ol",qe,[o("li",null,[o("b",null,s(e(t).t("zones.form.kubernetes.helm.step1Description")),1),n(),l($,{id:"zone-kubernetes-create-namespace",class:"mt-2",code:e(t).t("zones.form.kubernetes.helm.step1Command"),language:"bash"},null,8,["code"])]),n(),o("li",null,[o("b",null,s(e(t).t("zones.form.kubernetes.helm.step2Description")),1),n(),l($,{id:"zone-kubernetes-add-charts-repo",class:"mt-2",code:e(t).t("zones.form.kubernetes.helm.step2Command"),language:"bash"},null,8,["code"])]),n(),o("li",null,[o("b",null,s(e(t).t("zones.form.kubernetes.helm.step3Description")),1),n(),l($,{id:"zone-kubernetes-repo-update",class:"mt-2",code:e(t).t("zones.form.kubernetes.helm.step3Command"),language:"bash"},null,8,["code"])])]),n(),o("h3",Le,[Ae,n(" "+s(e(t).t("zones.form.kubernetes.secret.title")),1)]),n(),o("p",null,s(e(t).t("zones.form.kubernetes.secret.createSecretDescription")),1),n(),l($,{id:"zone-kubernetes-create-secret",class:"mt-4",code:I.value,language:"bash"},null,8,["code"]),n(),o("h3",De,[Be,n(" "+s(e(t).t("zones.form.kubernetes.connectZone.title")),1)]),n(),o("p",null,s(e(t).t("zones.form.kubernetes.connectZone.configDescription")),1),n(),o("span",Re,s(e(t).t("zones.form.kubernetes.connectZone.configFileName")),1),n(),l($,{id:"zone-kubernetes-config-code-block","data-testid":"zone-kubernetes-config",code:h.value,language:"yaml"},null,8,["code"]),n(),o("p",Ue,s(e(t).t("zones.form.kubernetes.connectZone.connectDescription")),1),n(),l($,{id:"zone-kubernetes-command-code-block",class:"mt-4",code:e(t).t("zones.form.kubernetes.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Fe={class:"form-step-title"},Oe=o("span",{class:"form-step-number"},"1",-1),Pe={class:"form-step-title"},Me=o("span",{class:"form-step-number"},"2",-1),Xe={class:"field-group-label mt-4"},Ge={class:"mt-4"},He=A({__name:"ZoneCreateUniversalInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},token:{type:String,required:!0}},setup(y){const t=x(),v=Q(),_=y,I=w(()=>t.t("zones.form.universal.saveToken.saveTokenCommand",{token:_.token}).trim()),h=w(()=>{const c={zoneName:_.zoneName,globalKdsAddress:_.globalKdsAddress};return typeof v.params.virtualControlPlaneId=="string"&&(c.controlPlaneId=v.params.virtualControlPlaneId),t.t("zones.form.universal.connectZone.config",c).trim()});return(c,E)=>(r(),d("div",null,[o("h3",Fe,[Oe,n(" "+s(e(t).t("zones.form.universal.saveToken.title")),1)]),n(),o("p",null,s(e(t).t("zones.form.universal.saveToken.saveTokenDescription")),1),n(),l($,{id:"zone-kubernetes-token",class:"mt-4",code:I.value,language:"bash"},null,8,["code"]),n(),o("h3",Pe,[Me,n(" "+s(e(t).t("zones.form.universal.connectZone.title")),1)]),n(),o("p",null,s(e(t).t("zones.form.universal.connectZone.configDescription")),1),n(),o("span",Xe,s(e(t).t("zones.form.universal.connectZone.configFileName")),1),n(),l($,{id:"zone-universal-config-code-block","data-testid":"zone-universal-config",class:"mt-4",code:h.value,language:"yaml"},null,8,["code"]),n(),o("p",Ge,s(e(t).t("zones.form.universal.connectZone.connectDescription")),1),n(),l($,{id:"zone-universal-connect-command-code-block",class:"mt-4",code:e(t).t("zones.form.universal.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),je={class:"form-wrapper"},Qe={key:1},We={key:2},Je={class:"form"},Ye={class:"form-header"},en={class:"form-title"},nn={class:"text-gradient"},tn={key:0},on={key:0},sn={class:"fact-list"},an={class:"form-section"},rn={class:"form-section__header"},ln={class:"form-section-title"},cn={key:0},un={class:"form-section__content"},dn={class:"form-section","data-testid":"connect-zone-instructions"},mn={class:"form-section__header"},pn={class:"form-section-title"},_n={key:0},fn={class:"form-section__content"},vn={class:"field-group-list"},bn={class:"field-group"},zn={class:"field-group-label"},hn={class:"radio-button-group"},gn={class:"field-group"},kn={class:"field-group-label"},yn={class:"radio-button-group"},En={class:"field-group"},Cn={class:"field-group-label"},$n={class:"radio-button-group"},In={class:"form-section"},wn={class:"form-section__header"},Kn={class:"form-section-title"},Vn={key:0},Sn={class:"form-section__content"},Nn={class:"form-section"},Tn={class:"form-section__header"},Zn={class:"form-section-title"},qn={key:0},Ln={class:"form-section__content"},An={class:"mt-2"},Dn=A({__name:"CreateView",setup(y){const{t,tm:v}=x(),_=ve(),I=be(),h=/^(?![-0-9])[a-z0-9-]{1,63}$/,c=f(null),E=f(!1),S=f(!1),m=f(null),g=f(null),N=f(!1),T=f(null),b=f(""),K=f("kubernetes"),D=f(!0),B=f(!0),Z=w(()=>c.value!==null&&c.value.token?c.value.token:""),W=w(()=>Z.value!==""?window.btoa(Z.value):""),J=w(()=>b.value===""||E.value||c.value!==null),R=w(()=>{if(g.value!==null)return g.value;if(m.value instanceof U){const p=m.value.invalidParameters.find(i=>i.field==="name");if(p!==void 0)return p.reason}return null});async function Y(){E.value=!0,m.value=null;try{if(!F(b.value))return;c.value=await _.createZone({name:b.value})}catch(p){p instanceof Error?m.value=p:console.error(p)}finally{E.value=!1}}function F(p){const i=h.test(p);return i?g.value=null:g.value=t("zones.create.invalidNameError"),i}async function ee(){N.value=!1,T.value=null;try{const p=await _.getZoneOverview({name:b.value}),i=ye(p.zoneInsight);N.value=i==="online"}catch(p){p instanceof Error?T.value=p:console.error(p)}}function O(){S.value=!S.value}return(p,i)=>{const ne=z("RouteTitle"),q=z("KButton"),te=z("KAlert"),oe=z("KLabel"),se=z("KInput"),P=z("KRadio"),M=z("KInputSwitch"),ae=z("DataSource"),re=z("KCard"),le=z("KModal"),ie=z("AppView"),ce=z("RouteView");return r(),k(ce,{name:"zone-create-view",attrs:{class:"is-fullscreen"}},{default:a(()=>[l(ie,{fullscreen:!0,breadcrumbs:[]},{title:a(()=>[o("h1",null,[l(ne,{title:e(t)("zones.routes.create.title"),render:!0},null,8,["title"])])]),actions:a(()=>[Z.value===""||N.value?(r(),k(q,{key:0,appearance:"outline","data-testid":"exit-button",to:{name:"zone-cp-list-view"}},{default:a(()=>[n(s(e(t)("zones.form.exit")),1)]),_:1})):(r(),k(q,{key:1,appearance:"outline","data-testid":"exit-button",onClick:O},{default:a(()=>[n(s(e(t)("zones.form.exit")),1)]),_:1}))]),default:a(()=>[n(),n(),o("div",je,[m.value!==null?(r(),k(te,{key:0,appearance:"danger",class:"mb-4","dismiss-type":"icon","data-testid":"create-zone-error"},{alertMessage:a(()=>[m.value instanceof e(U)&&[409,500].includes(m.value.status)?(r(),d(L,{key:0},[o("p",null,s(e(t)(`zones.create.status_error.${m.value.status}.title`,{name:b.value})),1),n(),o("p",null,s(e(t)(`zones.create.status_error.${m.value.status}.description`)),1)],64)):m.value instanceof e(U)?(r(),d("p",Qe,s(e(t)("common.error_state.api_error",{status:m.value.status,title:m.value.detail})),1)):(r(),d("p",We,s(e(t)("common.error_state.default_error")),1))]),_:1})):C("",!0),n(),l(re,{class:"form-card"},{body:a(()=>[o("div",Je,[o("div",Ye,[o("div",null,[o("h1",en,[o("span",nn,s(e(t)("zones.form.title")),1)]),n(),e(t)("zones.form.description")!==" "?(r(),d("p",tn,s(e(t)("zones.form.description")),1)):C("",!0)]),n(),e(v)("zones.form.facts").length>0?(r(),d("div",on,[o("ul",sn,[(r(!0),d(L,null,ze(e(v)("zones.form.facts"),(u,ue)=>(r(),d("li",{key:ue,class:"fact-list__item"},[l(e(he),{color:e(j)},null,8,["color"]),n(" "+s(u),1)]))),128))])])):C("",!0)]),n(),o("div",an,[o("div",rn,[o("h2",ln,s(e(t)("zones.form.section.name.title")),1),n(),e(t)("zones.form.section.name.description")!==" "?(r(),d("p",cn,s(e(t)("zones.form.section.name.description")),1)):C("",!0)]),n(),o("div",un,[o("div",null,[l(oe,{for:"zone-name",required:"","tooltip-attributes":{placement:"right"}},{tooltip:a(()=>[n(s(e(t)("zones.form.name_tooltip")),1)]),default:a(()=>[n(s(e(t)("zones.form.nameLabel"))+" ",1)]),_:1}),n(),l(se,{id:"zone-name",modelValue:b.value,"onUpdate:modelValue":i[0]||(i[0]=u=>b.value=u),type:"text",name:"zone-name","data-testid":"name-input","data-test-error-type":R.value!==null?"invalid-dns-name":void 0,"has-error":R.value!==null,"error-message":R.value??void 0,disabled:c.value!==null,onBlur:i[1]||(i[1]=u=>F(b.value))},null,8,["modelValue","data-test-error-type","has-error","error-message","disabled"])]),n(),l(q,{appearance:"primary",class:"mt-4",disabled:J.value,"data-testid":"create-zone-button",onClick:Y},{default:a(()=>[E.value?(r(),k(e(G),{key:0,color:e(H),size:e(X)},null,8,["color","size"])):(r(),k(e(ge),{key:1,size:e(X)},null,8,["size"])),n(" "+s(e(t)("zones.form.createZoneButtonLabel")),1)]),_:1},8,["disabled"])])]),n(),c.value!==null?(r(),d(L,{key:0},[o("div",dn,[o("div",mn,[o("h2",pn,s(e(t)("zones.form.section.configuration.title")),1),n(),e(t)("zones.form.section.configuration.description")!==" "?(r(),d("p",_n,s(e(t)("zones.form.section.configuration.description")),1)):C("",!0)]),n(),o("div",fn,[o("div",vn,[o("div",bn,[o("span",zn,s(e(t)("zones.form.environmentLabel"))+` *
                        `,1),n(),o("div",hn,[l(P,{id:"zone-environment-universal",modelValue:K.value,"onUpdate:modelValue":i[2]||(i[2]=u=>K.value=u),"selected-value":"universal",name:"zone-environment","data-testid":"environment-universal-radio-button"},{default:a(()=>[n(s(e(t)("zones.form.universalLabel")),1)]),_:1},8,["modelValue"]),n(),l(P,{id:"zone-environment-kubernetes",modelValue:K.value,"onUpdate:modelValue":i[3]||(i[3]=u=>K.value=u),"selected-value":"kubernetes",name:"zone-environment","data-testid":"environment-kubernetes-radio-button"},{default:a(()=>[n(s(e(t)("zones.form.kubernetesLabel")),1)]),_:1},8,["modelValue"])])]),n(),K.value==="kubernetes"?(r(),d(L,{key:0},[o("div",gn,[o("span",kn,s(e(t)("zones.form.zoneIngressLabel"))+` *
                          `,1),n(),o("div",yn,[l(M,{id:"zone-ingress-enabled",modelValue:D.value,"onUpdate:modelValue":i[4]||(i[4]=u=>D.value=u),"data-testid":"ingress-input-switch"},{label:a(()=>[n(s(e(t)("zones.form.zoneIngressEnabledLabel")),1)]),_:1},8,["modelValue"])])]),n(),o("div",En,[o("span",Cn,s(e(t)("zones.form.zoneEgressLabel"))+` *
                          `,1),n(),o("div",$n,[l(M,{id:"zone-egress-enabled",modelValue:B.value,"onUpdate:modelValue":i[5]||(i[5]=u=>B.value=u),"data-testid":"egress-input-switch"},{label:a(()=>[n(s(e(t)("zones.form.zoneEgressEnabledLabel")),1)]),_:1},8,["modelValue"])])])],64)):C("",!0)])])]),n(),o("div",In,[o("div",wn,[o("h2",Kn,s(e(t)("zones.form.section.connect_zone.title")),1),n(),e(t)("zones.form.section.connect_zone.description")!==" "?(r(),d("p",Vn,s(e(t)("zones.form.section.connect_zone.description")),1)):C("",!0)]),n(),o("div",Sn,[l(ae,{src:"/control-plane/addresses"},{default:a(({data:u})=>[typeof u<"u"?(r(),d(L,{key:0},[K.value==="universal"?(r(),k(He,{key:0,"zone-name":b.value,token:Z.value,"global-kds-address":u.kds},null,8,["zone-name","token","global-kds-address"])):(r(),k(xe,{key:1,"zone-name":b.value,"zone-ingress-enabled":D.value,"zone-egress-enabled":B.value,token:Z.value,"base64-encoded-token":W.value,"global-kds-address":u.kds},null,8,["zone-name","zone-ingress-enabled","zone-egress-enabled","token","base64-encoded-token","global-kds-address"]))],64)):C("",!0)]),_:1})])]),n(),o("div",Nn,[o("div",Tn,[o("h2",Zn,s(e(t)("zones.form.section.scanner.title")),1),n(),e(t)("zones.form.section.scanner.description")!==" "?(r(),d("p",qn,s(e(t)("zones.form.section.scanner.description")),1)):C("",!0)]),n(),o("div",Ln,[l(Ie,{"loader-function":ee,"has-error":T.value!==null,"can-complete":N.value,"data-testid":"zone-connected-scanner"},{"loading-title":a(()=>[n(s(e(t)("zones.form.scan.waitTitle")),1)]),"complete-title":a(()=>[n(s(e(t)("zones.form.scan.completeTitle")),1)]),"complete-content":a(()=>[o("p",null,s(e(t)("zones.form.scan.completeDescription",{name:b.value})),1),n(),o("p",An,[l(q,{appearance:"primary",to:{name:"zone-cp-detail-view",params:{zone:b.value}}},{default:a(()=>[n(s(e(t)("zones.form.scan.completeButtonLabel",{name:b.value})),1)]),_:1},8,["to"])])]),"error-title":a(()=>[o("h3",null,s(e(t)("zones.form.scan.errorTitle")),1)]),"error-content":a(()=>[o("p",null,s(e(t)("zones.form.scan.errorDescription")),1)]),_:1},8,["has-error","can-complete"])])])],64)):C("",!0)])]),_:1})]),n(),l(le,{"is-visible":S.value,title:e(t)("zones.form.confirm_modal.title"),"data-testid":"confirm-exit-modal",onCanceled:O,onProceed:i[6]||(i[6]=u=>e(I).push({name:"zone-cp-list-view"}))},{"header-content":a(()=>[n(s(e(t)("zones.form.confirm_modal.title")),1)]),"body-content":a(()=>[n(s(e(t)("zones.form.confirm_modal.body")),1)]),"action-buttons":a(()=>[l(q,{appearance:"primary",to:{name:"zone-cp-list-view"},"data-testid":"confirm-exit-button"},{default:a(()=>[n(s(e(t)("zones.form.confirm_modal.action_button")),1)]),_:1})]),_:1},8,["is-visible","title"])]),_:1})]),_:1})}}});const xn=ke(Dn,[["__scopeId","data-v-abc1af56"]]);export{xn as default};
